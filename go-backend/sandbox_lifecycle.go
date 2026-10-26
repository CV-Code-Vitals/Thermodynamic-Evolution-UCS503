package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"
	"github.com/google/uuid"
	"github.com/saksham/self-healing-pipeline/agent"
)

const (
	sandboxLifecycleRecordID  = "sandbox-lifecycle"
	sandboxLifecycleTable     = "sandbox-lifecycle"
	sandboxIdleDuration       = 10 * time.Minute
	sandboxLeaseGrace         = time.Minute
	sandboxStatePollInterval  = 100 * time.Millisecond
	sandboxBrokerPollInterval = time.Second
	sandboxDynamoRetryLimit   = 12
)

type SandboxState string

const (
	SandboxStateStarting SandboxState = "starting"
	SandboxStateRunning  SandboxState = "running"
	SandboxStateStopping SandboxState = "stopping"
	SandboxStateIdle     SandboxState = "idle"
)

var (
	ErrSandboxLifecycleStopping      = errors.New("sandbox lifecycle is stopping")
	ErrSandboxLifecycleNotDue        = errors.New("sandbox idle deadline is not due")
	ErrSandboxLifecycleRecordMissing = errors.New("sandbox lifecycle record is missing")
)

// The single-item contract is shared with the scheduled watchdog:
// id (string), state (string), active_count (number), idle_until_epoch (number),
// and leases (map<string, number>) containing per-lease expiry epoch seconds.
type SandboxLifecycleRecord struct {
	ID             string
	State          SandboxState
	ActiveCount    int64
	IdleUntilEpoch int64
	Leases         map[string]int64
}

func copyLeaseMap(source map[string]int64) map[string]int64 {
	result := make(map[string]int64, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

type SandboxLifecycleStore interface {
	Ensure(context.Context) error
	Get(context.Context) (SandboxLifecycleRecord, error)
	AcquireLease(context.Context, string, int64) (SandboxState, bool, error)
	MarkRunning(context.Context) error
	MarkStartupFailed(context.Context) error
	ReleaseLease(context.Context, string, int64) error
	ReconcileStaleLeases(context.Context, int64) (SandboxLifecycleRecord, error)
	RecoverStarting(context.Context, bool, int64) error
	ClaimStop(context.Context, int64) error
	MarkIdleAfterStop(context.Context) error
	AbortStop(context.Context) error
}

type SandboxInstanceClient interface {
	DescribeInstance(context.Context, string) (string, error)
	StartInstance(context.Context, string) error
	WaitInstanceRunning(context.Context, string) error
	StopInstance(context.Context, string) error
	WaitInstanceStopped(context.Context, string) error
}

type SandboxBrokerHealthChecker interface {
	CheckAvailable(context.Context) error
}

type SandboxLifecycleController interface {
	Acquire(context.Context) (func() error, error)
}

type SandboxLifecycle struct {
	store        SandboxLifecycleStore
	instance     SandboxInstanceClient
	broker       SandboxBrokerHealthChecker
	instanceID   string
	readyTimeout time.Duration
	now          func() time.Time

	timerMu   sync.Mutex
	idleTimer *time.Timer
}

func NewSandboxLifecycle(
	store SandboxLifecycleStore,
	instance SandboxInstanceClient,
	broker SandboxBrokerHealthChecker,
	instanceID string,
	readyTimeout time.Duration,
) (*SandboxLifecycle, error) {
	if store == nil || instance == nil || broker == nil {
		return nil, errors.New("sandbox lifecycle requires a lease store, EC2 client, and broker health checker")
	}
	if !validSandboxInstanceID(instanceID) {
		return nil, errors.New("SANDBOX_INSTANCE_ID must be a valid EC2 instance ID")
	}
	if readyTimeout <= 0 {
		return nil, errors.New("sandbox readiness timeout must be positive")
	}
	return &SandboxLifecycle{
		store:        store,
		instance:     instance,
		broker:       broker,
		instanceID:   instanceID,
		readyTimeout: readyTimeout,
		now:          time.Now,
	}, nil
}

func validSandboxInstanceID(value string) bool {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "i-") {
		return false
	}
	digits := value[2:]
	if len(digits) != 8 && len(digits) != 17 {
		return false
	}
	for _, char := range digits {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

func (l *SandboxLifecycle) Acquire(ctx context.Context) (func() error, error) {
	if err := l.store.Ensure(ctx); err != nil {
		return nil, fmt.Errorf("ensure sandbox lifecycle record: %w", err)
	}

	leaseExpiry := func() int64 {
		return l.now().Add(maxEvolutionDuration + l.readyTimeout + sandboxLeaseGrace).Unix()
	}
	var leaseID string
	var shouldStart bool
	for {
		leaseID = uuid.NewString()
		state, startOwner, err := l.store.AcquireLease(ctx, leaseID, leaseExpiry())
		if errors.Is(err, ErrSandboxLifecycleStopping) {
			timer := time.NewTimer(sandboxStatePollInterval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, fmt.Errorf("wait for sandbox stop to finish: %w", ctx.Err())
			case <-timer.C:
				continue
			}
		}
		if err != nil {
			return nil, fmt.Errorf("acquire sandbox work lease: %w", err)
		}
		shouldStart = startOwner
		if state != SandboxStateStarting && state != SandboxStateRunning {
			cleanupErr := l.releaseLease(leaseID)
			return nil, errors.Join(
				fmt.Errorf("sandbox lease acquired in unexpected state %q", state),
				cleanupErr,
			)
		}
		break
	}
	l.cancelIdleTimer()

	readyCtx, cancel := context.WithTimeout(ctx, l.readyTimeout)
	err := l.ensureReady(readyCtx, shouldStart)
	cancel()
	if err == nil {
		err = l.store.MarkRunning(ctx)
		if err != nil {
			err = fmt.Errorf("mark sandbox running after readiness: %w", err)
		}
	}
	if err != nil {
		cleanupErr := l.releaseLease(leaseID)
		return nil, errors.Join(err, cleanupErr)
	}

	var once sync.Once
	var releaseErr error
	return func() error {
		once.Do(func() {
			releaseErr = l.releaseLease(leaseID)
		})
		return releaseErr
	}, nil
}

func (l *SandboxLifecycle) ensureReady(ctx context.Context, startOwner bool) error {
	if startOwner {
		state, err := l.instance.DescribeInstance(ctx, l.instanceID)
		if err != nil {
			return fmt.Errorf("describe configured sandbox instance %s: %w", l.instanceID, err)
		}
		switch state {
		case "stopped":
			if err := l.instance.StartInstance(ctx, l.instanceID); err != nil {
				return fmt.Errorf("start configured sandbox instance %s: %w", l.instanceID, err)
			}
		case "running", "pending":
		case "stopping":
			if err := l.instance.WaitInstanceStopped(ctx, l.instanceID); err != nil {
				return fmt.Errorf("wait for configured sandbox instance %s to stop: %w", l.instanceID, err)
			}
			if err := l.instance.StartInstance(ctx, l.instanceID); err != nil {
				return fmt.Errorf("start configured sandbox instance %s: %w", l.instanceID, err)
			}
		default:
			return fmt.Errorf("configured sandbox instance %s is in unsupported EC2 state %q", l.instanceID, state)
		}
	}
	if err := l.instance.WaitInstanceRunning(ctx, l.instanceID); err != nil {
		return fmt.Errorf("wait for configured sandbox instance %s to run: %w", l.instanceID, err)
	}

	var lastErr error
	for {
		if err := l.broker.CheckAvailable(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}
		timer := time.NewTimer(sandboxBrokerPollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("wait for authenticated sandbox broker health: %w (last health check: %v)", ctx.Err(), lastErr)
		case <-timer.C:
		}
	}
}

func (l *SandboxLifecycle) releaseLease(leaseID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	idleUntil := l.now().Add(sandboxIdleDuration).Unix()
	if err := l.store.ReleaseLease(ctx, leaseID, idleUntil); err != nil {
		return fmt.Errorf("release sandbox work lease: %w", err)
	}
	record, err := l.store.Get(ctx)
	if err != nil {
		return fmt.Errorf("read sandbox lifecycle after lease release: %w", err)
	}
	if record.ActiveCount == 0 && record.State == SandboxStateIdle && record.IdleUntilEpoch > 0 {
		l.scheduleIdleStop(time.Unix(record.IdleUntilEpoch, 0))
	}
	return nil
}

func (l *SandboxLifecycle) cancelIdleTimer() {
	l.timerMu.Lock()
	defer l.timerMu.Unlock()
	if l.idleTimer != nil {
		l.idleTimer.Stop()
		l.idleTimer = nil
	}
}

func (l *SandboxLifecycle) scheduleIdleStop(deadline time.Time) {
	l.timerMu.Lock()
	defer l.timerMu.Unlock()
	if l.idleTimer != nil {
		l.idleTimer.Stop()
	}
	delay := time.Until(deadline)
	if delay < 0 {
		delay = 0
	}
	l.idleTimer = time.AfterFunc(delay, func() {
		ctx, cancel := context.WithTimeout(context.Background(), l.readyTimeout+2*time.Minute)
		defer cancel()
		if err := l.Reconcile(ctx, l.now()); err != nil {
			log.Printf("sandbox idle-stop reconciliation failed: %v", err)
		}
	})
}

// Reconcile is shared by the local idle timer and the scheduled SAM watchdog.
func (l *SandboxLifecycle) Reconcile(ctx context.Context, now time.Time) error {
	record, err := l.store.ReconcileStaleLeases(ctx, now.Unix())
	if err != nil {
		return fmt.Errorf("reconcile stale sandbox leases: %w", err)
	}

	if record.State == SandboxStateStarting && record.ActiveCount == 0 {
		instanceState, err := l.instance.DescribeInstance(ctx, l.instanceID)
		if err != nil {
			return fmt.Errorf("describe sandbox while recovering startup: %w", err)
		}
		instanceRunning := instanceState == "running"
		if instanceState == "pending" {
			if err := l.instance.WaitInstanceRunning(ctx, l.instanceID); err != nil {
				return fmt.Errorf("wait for sandbox during startup recovery: %w", err)
			}
			instanceRunning = true
		} else if instanceState == "stopping" {
			if err := l.instance.WaitInstanceStopped(ctx, l.instanceID); err != nil {
				return fmt.Errorf("wait for sandbox stop during startup recovery: %w", err)
			}
		}
		if err := l.store.RecoverStarting(ctx, instanceRunning, now.Unix()); err != nil {
			return fmt.Errorf("recover abandoned sandbox startup: %w", err)
		}
		record, err = l.store.Get(ctx)
		if err != nil {
			return fmt.Errorf("read sandbox after startup recovery: %w", err)
		}
	}
	if record.State != SandboxStateIdle || record.ActiveCount != 0 ||
		record.IdleUntilEpoch == 0 || record.IdleUntilEpoch > now.Unix() {
		return nil
	}
	if err := l.store.ClaimStop(ctx, now.Unix()); err != nil {
		if errors.Is(err, ErrSandboxLifecycleNotDue) {
			return nil
		}
		return fmt.Errorf("claim idle sandbox stop: %w", err)
	}
	if err := l.stopClaimedInstance(ctx); err != nil {
		rollbackErr := l.store.AbortStop(ctx)
		return errors.Join(err, wrapNonNil("restore lifecycle state after stop failure", rollbackErr))
	}
	if err := l.store.MarkIdleAfterStop(ctx); err != nil {
		return fmt.Errorf("record stopped sandbox instance: %w", err)
	}
	return nil
}

func wrapNonNil(prefix string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", prefix, err)
}

func (l *SandboxLifecycle) stopClaimedInstance(ctx context.Context) error {
	state, err := l.instance.DescribeInstance(ctx, l.instanceID)
	if err != nil {
		return fmt.Errorf("describe configured sandbox before stop: %w", err)
	}
	switch state {
	case "stopped", "terminated":
		return nil
	case "pending":
		if err := l.instance.WaitInstanceRunning(ctx, l.instanceID); err != nil {
			return fmt.Errorf("wait for sandbox before stop: %w", err)
		}
	case "running":
		if err := l.instance.StopInstance(ctx, l.instanceID); err != nil {
			return fmt.Errorf("stop configured sandbox instance %s: %w", l.instanceID, err)
		}
	case "stopping":
	default:
		return fmt.Errorf("configured sandbox instance %s is in unsupported EC2 state %q", l.instanceID, state)
	}
	if state == "pending" {
		if err := l.instance.StopInstance(ctx, l.instanceID); err != nil {
			return fmt.Errorf("stop configured sandbox instance %s: %w", l.instanceID, err)
		}
	}
	if err := l.instance.WaitInstanceStopped(ctx, l.instanceID); err != nil {
		return fmt.Errorf("wait for configured sandbox instance %s to stop: %w", l.instanceID, err)
	}
	return nil
}

type dynamoSandboxLifecycleAPI interface {
	GetItem(context.Context, *dynamodb.GetItemInput, ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	UpdateItem(context.Context, *dynamodb.UpdateItemInput, ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

type DynamoSandboxLifecycleStore struct {
	client    dynamoSandboxLifecycleAPI
	tableName string
}

func NewDynamoSandboxLifecycleStore(client dynamoSandboxLifecycleAPI, tableName string) (*DynamoSandboxLifecycleStore, error) {
	if client == nil {
		return nil, errors.New("DynamoDB lifecycle client is required")
	}
	if strings.TrimSpace(tableName) != sandboxLifecycleTable {
		return nil, fmt.Errorf("SANDBOX_LIFECYCLE_TABLE must be %q", sandboxLifecycleTable)
	}
	return &DynamoSandboxLifecycleStore{client: client, tableName: tableName}, nil
}

func (s *DynamoSandboxLifecycleStore) key() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: sandboxLifecycleRecordID},
	}
}

func (s *DynamoSandboxLifecycleStore) Ensure(ctx context.Context) error {
	_, err := s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(s.tableName),
		ConditionExpression: aws.String("attribute_not_exists(#id)"),
		ExpressionAttributeNames: map[string]string{
			"#id": "id",
		},
		Item: map[string]types.AttributeValue{
			"id":               &types.AttributeValueMemberS{Value: sandboxLifecycleRecordID},
			"state":            &types.AttributeValueMemberS{Value: string(SandboxStateIdle)},
			"active_count":     &types.AttributeValueMemberN{Value: "0"},
			"idle_until_epoch": &types.AttributeValueMemberN{Value: "0"},
			"leases":           &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}},
		},
	})
	if isConditionalCheckFailure(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("initialize DynamoDB sandbox lifecycle record: %w", err)
	}
	return nil
}

func (s *DynamoSandboxLifecycleStore) Get(ctx context.Context) (SandboxLifecycleRecord, error) {
	output, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(s.tableName),
		Key:            s.key(),
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return SandboxLifecycleRecord{}, fmt.Errorf("read DynamoDB sandbox lifecycle record: %w", err)
	}
	if len(output.Item) == 0 {
		return SandboxLifecycleRecord{}, ErrSandboxLifecycleRecordMissing
	}
	record, err := decodeSandboxLifecycleRecord(output.Item)
	if err != nil {
		return SandboxLifecycleRecord{}, fmt.Errorf("decode DynamoDB sandbox lifecycle record: %w", err)
	}
	return record, nil
}

func decodeSandboxLifecycleRecord(item map[string]types.AttributeValue) (SandboxLifecycleRecord, error) {
	record := SandboxLifecycleRecord{}
	if value, ok := item["id"].(*types.AttributeValueMemberS); ok {
		record.ID = value.Value
	}
	state, ok := item["state"].(*types.AttributeValueMemberS)
	if !ok {
		return record, errors.New("state must be a DynamoDB string")
	}
	record.State = SandboxState(state.Value)
	if record.State != SandboxStateIdle && record.State != SandboxStateStarting &&
		record.State != SandboxStateRunning && record.State != SandboxStateStopping {
		return record, fmt.Errorf("unsupported lifecycle state %q", state.Value)
	}
	var err error
	if record.ActiveCount, err = readDynamoInt(item["active_count"], "active_count"); err != nil {
		return record, err
	}
	if record.IdleUntilEpoch, err = readDynamoInt(item["idle_until_epoch"], "idle_until_epoch"); err != nil {
		return record, err
	}
	leases, ok := item["leases"].(*types.AttributeValueMemberM)
	if !ok {
		return record, errors.New("leases must be a DynamoDB map")
	}
	record.Leases = make(map[string]int64, len(leases.Value))
	for leaseID, expiry := range leases.Value {
		epoch, err := readDynamoInt(expiry, "leases."+leaseID)
		if err != nil {
			return record, err
		}
		record.Leases[leaseID] = epoch
	}
	if record.ID != sandboxLifecycleRecordID || record.ActiveCount < 0 {
		return record, errors.New("lifecycle record id or active_count is invalid")
	}
	return record, nil
}

func readDynamoInt(value types.AttributeValue, name string) (int64, error) {
	number, ok := value.(*types.AttributeValueMemberN)
	if !ok {
		return 0, fmt.Errorf("%s must be a DynamoDB number", name)
	}
	parsed, err := strconv.ParseInt(number.Value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	return parsed, nil
}

func (s *DynamoSandboxLifecycleStore) AcquireLease(ctx context.Context, leaseID string, expires int64) (SandboxState, bool, error) {
	for attempt := 0; attempt < sandboxDynamoRetryLimit; attempt++ {
		record, err := s.Get(ctx)
		if err != nil {
			return "", false, err
		}
		if record.State == SandboxStateStopping {
			return record.State, false, ErrSandboxLifecycleStopping
		}
		startOwner := record.State == SandboxStateIdle
		state := record.State
		if startOwner {
			state = SandboxStateStarting
		} else if state != SandboxStateStarting && state != SandboxStateRunning {
			return "", false, fmt.Errorf("cannot acquire lease in lifecycle state %q", state)
		}
		nextCount := record.ActiveCount + 1
		_, err = s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName:           aws.String(s.tableName),
			Key:                 s.key(),
			UpdateExpression:    aws.String("SET #state = :next_state, #active = :next_count, #idle = :zero, #leases.#lease = :expires"),
			ConditionExpression: aws.String("#state = :old_state AND #active = :old_count AND attribute_not_exists(#leases.#lease)"),
			ExpressionAttributeNames: map[string]string{
				"#state": "state", "#active": "active_count", "#idle": "idle_until_epoch",
				"#leases": "leases", "#lease": leaseID,
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":next_state": &types.AttributeValueMemberS{Value: string(state)},
				":old_state":  &types.AttributeValueMemberS{Value: string(record.State)},
				":next_count": &types.AttributeValueMemberN{Value: strconv.FormatInt(nextCount, 10)},
				":old_count":  &types.AttributeValueMemberN{Value: strconv.FormatInt(record.ActiveCount, 10)},
				":zero":       &types.AttributeValueMemberN{Value: "0"},
				":expires":    &types.AttributeValueMemberN{Value: strconv.FormatInt(expires, 10)},
			},
		})
		if isConditionalCheckFailure(err) {
			continue
		}
		if err != nil {
			return "", false, fmt.Errorf("conditionally acquire DynamoDB sandbox lease: %w", err)
		}
		return state, startOwner, nil
	}
	return "", false, errors.New("sandbox lease acquisition exceeded conditional retry limit")
}

func (s *DynamoSandboxLifecycleStore) MarkRunning(ctx context.Context) error {
	for attempt := 0; attempt < sandboxDynamoRetryLimit; attempt++ {
		record, err := s.Get(ctx)
		if err != nil {
			return err
		}
		if record.State == SandboxStateRunning || (record.State == SandboxStateIdle && record.ActiveCount == 0) {
			return nil
		}
		if record.State != SandboxStateStarting && record.State != SandboxStateIdle {
			return fmt.Errorf("cannot mark sandbox running from %q", record.State)
		}
		if record.ActiveCount == 0 {
			return nil
		}
		_, err = s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(s.tableName), Key: s.key(),
			UpdateExpression:    aws.String("SET #state = :running, #idle = :zero"),
			ConditionExpression: aws.String("#state = :old_state AND #active = :count"),
			ExpressionAttributeNames: map[string]string{
				"#state": "state", "#idle": "idle_until_epoch", "#active": "active_count",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":running":   &types.AttributeValueMemberS{Value: string(SandboxStateRunning)},
				":old_state": &types.AttributeValueMemberS{Value: string(record.State)},
				":count":     &types.AttributeValueMemberN{Value: strconv.FormatInt(record.ActiveCount, 10)},
				":zero":      &types.AttributeValueMemberN{Value: "0"},
			},
		})
		if isConditionalCheckFailure(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("mark DynamoDB sandbox lifecycle running: %w", err)
		}
		return nil
	}
	return errors.New("marking sandbox running exceeded conditional retry limit")
}

func (s *DynamoSandboxLifecycleStore) MarkStartupFailed(ctx context.Context) error {
	_, err := s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName), Key: s.key(),
		UpdateExpression:         aws.String("SET #state = :idle, #idle = :zero"),
		ConditionExpression:      aws.String("#state = :starting"),
		ExpressionAttributeNames: map[string]string{"#state": "state", "#idle": "idle_until_epoch"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":idle":     &types.AttributeValueMemberS{Value: string(SandboxStateIdle)},
			":starting": &types.AttributeValueMemberS{Value: string(SandboxStateStarting)},
			":zero":     &types.AttributeValueMemberN{Value: "0"},
		},
	})
	if isConditionalCheckFailure(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("mark DynamoDB sandbox startup failed: %w", err)
	}
	return nil
}

func (s *DynamoSandboxLifecycleStore) ReleaseLease(ctx context.Context, leaseID string, idleUntil int64) error {
	for attempt := 0; attempt < sandboxDynamoRetryLimit; attempt++ {
		record, err := s.Get(ctx)
		if err != nil {
			if errors.Is(err, ErrSandboxLifecycleRecordMissing) {
				return err
			}
			return err
		}
		expiry, exists := record.Leases[leaseID]
		if !exists {
			return nil
		}
		if record.ActiveCount <= 0 || record.State == SandboxStateStopping {
			return errors.New("sandbox lifecycle lease/count invariant violated during release")
		}
		nextCount := record.ActiveCount - 1
		nextState := record.State
		nextIdle := int64(0)
		if nextCount == 0 {
			nextState = SandboxStateIdle
			nextIdle = idleUntil
		}
		_, err = s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(s.tableName), Key: s.key(),
			UpdateExpression:    aws.String("SET #state = :next_state, #active = :next_count, #idle = :next_idle REMOVE #leases.#lease"),
			ConditionExpression: aws.String("#state = :old_state AND #active = :old_count AND #leases.#lease = :expiry"),
			ExpressionAttributeNames: map[string]string{
				"#state": "state", "#active": "active_count", "#idle": "idle_until_epoch",
				"#leases": "leases", "#lease": leaseID,
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":next_state": &types.AttributeValueMemberS{Value: string(nextState)},
				":old_state":  &types.AttributeValueMemberS{Value: string(record.State)},
				":next_count": &types.AttributeValueMemberN{Value: strconv.FormatInt(nextCount, 10)},
				":old_count":  &types.AttributeValueMemberN{Value: strconv.FormatInt(record.ActiveCount, 10)},
				":next_idle":  &types.AttributeValueMemberN{Value: strconv.FormatInt(nextIdle, 10)},
				":expiry":     &types.AttributeValueMemberN{Value: strconv.FormatInt(expiry, 10)},
			},
		})
		if isConditionalCheckFailure(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("conditionally release DynamoDB sandbox lease: %w", err)
		}
		return nil
	}
	return errors.New("sandbox lease release exceeded conditional retry limit")
}

func (s *DynamoSandboxLifecycleStore) ReconcileStaleLeases(ctx context.Context, now int64) (SandboxLifecycleRecord, error) {
	for attempt := 0; attempt < sandboxDynamoRetryLimit; attempt++ {
		record, err := s.Get(ctx)
		if err != nil {
			return SandboxLifecycleRecord{}, err
		}
		stale := make(map[string]int64)
		for leaseID, expiry := range record.Leases {
			if expiry <= now {
				stale[leaseID] = expiry
			}
		}
		if len(stale) == 0 {
			return record, nil
		}
		nextCount := record.ActiveCount - int64(len(stale))
		if nextCount < 0 {
			return SandboxLifecycleRecord{}, errors.New("sandbox active_count is smaller than the stale lease map")
		}
		names := map[string]string{
			"#state": "state", "#active": "active_count", "#idle": "idle_until_epoch",
			"#leases": "leases",
		}
		values := map[string]types.AttributeValue{
			":old_state":  &types.AttributeValueMemberS{Value: string(record.State)},
			":old_count":  &types.AttributeValueMemberN{Value: strconv.FormatInt(record.ActiveCount, 10)},
			":next_count": &types.AttributeValueMemberN{Value: strconv.FormatInt(nextCount, 10)},
			":now_idle":   &types.AttributeValueMemberN{Value: strconv.FormatInt(now+int64(sandboxIdleDuration/time.Second), 10)},
		}
		conditions := []string{"#state = :old_state", "#active = :old_count"}
		removePaths := make([]string, 0, len(stale))
		index := 0
		for leaseID, expiry := range stale {
			name := fmt.Sprintf("#lease%d", index)
			value := fmt.Sprintf(":expiry%d", index)
			names[name] = leaseID
			values[value] = &types.AttributeValueMemberN{Value: strconv.FormatInt(expiry, 10)}
			conditions = append(conditions, "#leases."+name+" = "+value)
			removePaths = append(removePaths, "#leases."+name)
			index++
		}
		updateExpression := "SET #active = :next_count"
		if nextCount == 0 && record.State != SandboxStateStarting {
			values[":idle_state"] = &types.AttributeValueMemberS{Value: string(SandboxStateIdle)}
			updateExpression += ", #state = :idle_state, #idle = :now_idle"
		}
		updateExpression += " REMOVE " + strings.Join(removePaths, ", ")
		_, err = s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(s.tableName), Key: s.key(),
			UpdateExpression:          aws.String(updateExpression),
			ConditionExpression:       aws.String(strings.Join(conditions, " AND ")),
			ExpressionAttributeNames:  names,
			ExpressionAttributeValues: values,
		})
		if isConditionalCheckFailure(err) {
			continue
		}
		if err != nil {
			return SandboxLifecycleRecord{}, fmt.Errorf("conditionally reconcile stale DynamoDB sandbox leases: %w", err)
		}
		return s.Get(ctx)
	}
	return SandboxLifecycleRecord{}, errors.New("stale sandbox lease reconciliation exceeded conditional retry limit")
}

func (s *DynamoSandboxLifecycleStore) RecoverStarting(ctx context.Context, instanceRunning bool, now int64) error {
	nextIdle := int64(0)
	if instanceRunning {
		nextIdle = now + int64(sandboxIdleDuration/time.Second)
	}
	_, err := s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName), Key: s.key(),
		UpdateExpression:         aws.String("SET #state = :idle, #idle = :deadline"),
		ConditionExpression:      aws.String("#state = :starting AND #active = :zero"),
		ExpressionAttributeNames: map[string]string{"#state": "state", "#idle": "idle_until_epoch", "#active": "active_count"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":idle":     &types.AttributeValueMemberS{Value: string(SandboxStateIdle)},
			":starting": &types.AttributeValueMemberS{Value: string(SandboxStateStarting)},
			":zero":     &types.AttributeValueMemberN{Value: "0"},
			":deadline": &types.AttributeValueMemberN{Value: strconv.FormatInt(nextIdle, 10)},
		},
	})
	if isConditionalCheckFailure(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("recover DynamoDB sandbox startup state: %w", err)
	}
	return nil
}

func (s *DynamoSandboxLifecycleStore) ClaimStop(ctx context.Context, now int64) error {
	_, err := s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName), Key: s.key(),
		UpdateExpression:         aws.String("SET #state = :stopping"),
		ConditionExpression:      aws.String("#state = :idle AND #active = :zero AND #idle > :zero AND #idle <= :now"),
		ExpressionAttributeNames: map[string]string{"#state": "state", "#active": "active_count", "#idle": "idle_until_epoch"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":stopping": &types.AttributeValueMemberS{Value: string(SandboxStateStopping)},
			":idle":     &types.AttributeValueMemberS{Value: string(SandboxStateIdle)},
			":zero":     &types.AttributeValueMemberN{Value: "0"},
			":now":      &types.AttributeValueMemberN{Value: strconv.FormatInt(now, 10)},
		},
	})
	if isConditionalCheckFailure(err) {
		return ErrSandboxLifecycleNotDue
	}
	if err != nil {
		return fmt.Errorf("conditionally claim DynamoDB sandbox stop: %w", err)
	}
	return nil
}

func (s *DynamoSandboxLifecycleStore) MarkIdleAfterStop(ctx context.Context) error {
	_, err := s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName), Key: s.key(),
		UpdateExpression:         aws.String("SET #state = :idle, #idle = :zero"),
		ConditionExpression:      aws.String("#state = :stopping AND #active = :zero"),
		ExpressionAttributeNames: map[string]string{"#state": "state", "#idle": "idle_until_epoch", "#active": "active_count"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":idle":     &types.AttributeValueMemberS{Value: string(SandboxStateIdle)},
			":stopping": &types.AttributeValueMemberS{Value: string(SandboxStateStopping)},
			":zero":     &types.AttributeValueMemberN{Value: "0"},
		},
	})
	if err != nil {
		return fmt.Errorf("record stopped DynamoDB sandbox state: %w", err)
	}
	return nil
}

func (s *DynamoSandboxLifecycleStore) AbortStop(ctx context.Context) error {
	_, err := s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName), Key: s.key(),
		UpdateExpression:         aws.String("SET #state = :idle"),
		ConditionExpression:      aws.String("#state = :stopping AND #active = :zero"),
		ExpressionAttributeNames: map[string]string{"#state": "state", "#active": "active_count"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":idle":     &types.AttributeValueMemberS{Value: string(SandboxStateIdle)},
			":stopping": &types.AttributeValueMemberS{Value: string(SandboxStateStopping)},
			":zero":     &types.AttributeValueMemberN{Value: "0"},
		},
	})
	if isConditionalCheckFailure(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("restore idle DynamoDB sandbox state after stop failure: %w", err)
	}
	return nil
}

func isConditionalCheckFailure(err error) bool {
	var apiError smithy.APIError
	return errors.As(err, &apiError) && apiError.ErrorCode() == "ConditionalCheckFailedException"
}

type ec2SandboxLifecycleAPI interface {
	DescribeInstances(context.Context, *ec2.DescribeInstancesInput, ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	StartInstances(context.Context, *ec2.StartInstancesInput, ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	StopInstances(context.Context, *ec2.StopInstancesInput, ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
}

type EC2SandboxInstanceClient struct {
	client ec2SandboxLifecycleAPI
}

func NewEC2SandboxInstanceClient(client ec2SandboxLifecycleAPI) (*EC2SandboxInstanceClient, error) {
	if client == nil {
		return nil, errors.New("EC2 lifecycle client is required")
	}
	return &EC2SandboxInstanceClient{client: client}, nil
}

func (c *EC2SandboxInstanceClient) DescribeInstance(ctx context.Context, instanceID string) (string, error) {
	output, err := c.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{InstanceIds: []string{instanceID}})
	if err != nil {
		return "", fmt.Errorf("EC2 DescribeInstances for configured instance %s: %w", instanceID, err)
	}
	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			if aws.ToString(instance.InstanceId) == instanceID {
				return string(instance.State.Name), nil
			}
		}
	}
	return "", fmt.Errorf("EC2 DescribeInstances did not return configured instance %s", instanceID)
}

func (c *EC2SandboxInstanceClient) StartInstance(ctx context.Context, instanceID string) error {
	_, err := c.client.StartInstances(ctx, &ec2.StartInstancesInput{InstanceIds: []string{instanceID}})
	if err != nil {
		return fmt.Errorf("EC2 StartInstances for configured instance %s: %w", instanceID, err)
	}
	return nil
}

func (c *EC2SandboxInstanceClient) WaitInstanceRunning(ctx context.Context, instanceID string) error {
	waiter := ec2.NewInstanceRunningWaiter(c.client)
	if err := waiter.Wait(ctx, &ec2.DescribeInstancesInput{InstanceIds: []string{instanceID}}, 2*time.Minute); err != nil {
		return fmt.Errorf("EC2 wait until configured instance %s is running: %w", instanceID, err)
	}
	return nil
}

func (c *EC2SandboxInstanceClient) StopInstance(ctx context.Context, instanceID string) error {
	_, err := c.client.StopInstances(ctx, &ec2.StopInstancesInput{InstanceIds: []string{instanceID}})
	if err != nil {
		return fmt.Errorf("EC2 StopInstances for configured instance %s: %w", instanceID, err)
	}
	return nil
}

func (c *EC2SandboxInstanceClient) WaitInstanceStopped(ctx context.Context, instanceID string) error {
	waiter := ec2.NewInstanceStoppedWaiter(c.client)
	if err := waiter.Wait(ctx, &ec2.DescribeInstancesInput{InstanceIds: []string{instanceID}}, 2*time.Minute); err != nil {
		return fmt.Errorf("EC2 wait until configured instance %s is stopped: %w", instanceID, err)
	}
	return nil
}

func sandboxLifecycleFromEnv(ctx context.Context) (*SandboxLifecycle, error) {
	instanceID := strings.TrimSpace(os.Getenv("SANDBOX_INSTANCE_ID"))
	tableName := strings.TrimSpace(os.Getenv("SANDBOX_LIFECYCLE_TABLE"))
	if instanceID == "" || tableName == "" {
		return nil, errors.New("SANDBOX_INSTANCE_ID and SANDBOX_LIFECYCLE_TABLE must be configured")
	}
	readyTimeoutSeconds := 180
	if raw := strings.TrimSpace(os.Getenv("SANDBOX_READY_TIMEOUT_SECONDS")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 900 {
			return nil, errors.New("SANDBOX_READY_TIMEOUT_SECONDS must be between 1 and 900")
		}
		readyTimeoutSeconds = value
	}
	if strings.TrimSpace(os.Getenv("SANDBOX_SERVICE_URL")) == "" ||
		strings.TrimSpace(os.Getenv("SANDBOX_SERVICE_TOKEN")) == "" ||
		strings.TrimSpace(os.Getenv("SANDBOX_TLS_CA_FILE")) == "" {
		return nil, errors.New("SANDBOX_SERVICE_URL, SANDBOX_SERVICE_TOKEN, and SANDBOX_TLS_CA_FILE are required for authenticated broker readiness")
	}
	awsConfig, err := config.LoadDefaultConfig(
		ctx,
		config.WithUseDualStackEndpoint(aws.DualStackEndpointStateEnabled),
	)
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration for sandbox lifecycle: %w", err)
	}
	store, err := NewDynamoSandboxLifecycleStore(dynamodb.NewFromConfig(awsConfig), tableName)
	if err != nil {
		return nil, err
	}
	instance, err := NewEC2SandboxInstanceClient(ec2.NewFromConfig(awsConfig))
	if err != nil {
		return nil, err
	}
	broker := agent.NewRemoteBrokerEvaluator(
		os.Getenv("SANDBOX_SERVICE_URL"),
		os.Getenv("SANDBOX_SERVICE_TOKEN"),
	)
	return NewSandboxLifecycle(store, instance, broker, instanceID, time.Duration(readyTimeoutSeconds)*time.Second)
}
