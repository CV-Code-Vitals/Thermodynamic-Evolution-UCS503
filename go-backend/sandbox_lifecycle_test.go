package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type fakeSandboxStore struct {
	mu      sync.Mutex
	record  SandboxLifecycleRecord
	leases  map[string]int64
	ensured bool
}

func newFakeSandboxStore() *fakeSandboxStore {
	return &fakeSandboxStore{
		record: SandboxLifecycleRecord{
			ID:             sandboxLifecycleRecordID,
			State:          SandboxStateIdle,
			IdleUntilEpoch: 0,
		},
		leases: make(map[string]int64),
	}
}

func (s *fakeSandboxStore) Ensure(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensured = true
	return nil
}

func (s *fakeSandboxStore) Get(context.Context) (SandboxLifecycleRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.record.Leases = copyLeaseMap(s.leases)
	return s.record, nil
}

func (s *fakeSandboxStore) AcquireLease(_ context.Context, leaseID string, expires int64) (SandboxState, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.leases[leaseID]; exists {
		return "", false, errors.New("duplicate lease")
	}
	if s.record.State == SandboxStateStopping {
		return s.record.State, false, ErrSandboxLifecycleStopping
	}
	wasIdle := s.record.State == SandboxStateIdle
	if wasIdle {
		s.record.State = SandboxStateStarting
	}
	s.leases[leaseID] = expires
	s.record.ActiveCount++
	s.record.IdleUntilEpoch = 0
	s.record.Leases = copyLeaseMap(s.leases)
	return s.record.State, wasIdle, nil
}

func (s *fakeSandboxStore) MarkRunning(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.record.State != SandboxStateStarting {
		return fmt.Errorf("mark running from %s", s.record.State)
	}
	if s.record.ActiveCount == 0 {
		s.record.State = SandboxStateIdle
		return nil
	}
	s.record.State = SandboxStateRunning
	return nil
}

func (s *fakeSandboxStore) MarkStartupFailed(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.record.State == SandboxStateStarting {
		s.record.State = SandboxStateIdle
		s.record.IdleUntilEpoch = 0
	}
	return nil
}

func (s *fakeSandboxStore) ReleaseLease(_ context.Context, leaseID string, idleUntil int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.leases[leaseID]; !exists {
		return nil
	}
	delete(s.leases, leaseID)
	s.record.ActiveCount--
	s.record.Leases = copyLeaseMap(s.leases)
	if s.record.ActiveCount == 0 {
		s.record.State = SandboxStateIdle
		s.record.IdleUntilEpoch = idleUntil
	}
	return nil
}

func (s *fakeSandboxStore) ReconcileStaleLeases(_ context.Context, now int64) (SandboxLifecycleRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for leaseID, expires := range s.leases {
		if expires <= now {
			delete(s.leases, leaseID)
			s.record.ActiveCount--
		}
	}
	if s.record.ActiveCount == 0 && s.record.State == SandboxStateRunning {
		s.record.State = SandboxStateIdle
		if s.record.IdleUntilEpoch == 0 {
			s.record.IdleUntilEpoch = now + int64(sandboxIdleDuration/time.Second)
		}
	}
	s.record.Leases = copyLeaseMap(s.leases)
	return s.record, nil
}

func (s *fakeSandboxStore) RecoverStarting(_ context.Context, instanceRunning bool, now int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.record.State != SandboxStateStarting || s.record.ActiveCount != 0 {
		return nil
	}
	if instanceRunning {
		s.record.State = SandboxStateIdle
		s.record.IdleUntilEpoch = now + int64(sandboxIdleDuration/time.Second)
	} else {
		s.record.State = SandboxStateIdle
		s.record.IdleUntilEpoch = 0
	}
	return nil
}

func (s *fakeSandboxStore) ClaimStop(_ context.Context, now int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.record.State != SandboxStateIdle || s.record.ActiveCount != 0 || s.record.IdleUntilEpoch > now || s.record.IdleUntilEpoch == 0 {
		return ErrSandboxLifecycleNotDue
	}
	s.record.State = SandboxStateStopping
	return nil
}

func (s *fakeSandboxStore) MarkIdleAfterStop(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.record.State != SandboxStateStopping || s.record.ActiveCount != 0 {
		return errors.New("unsafe idle transition")
	}
	s.record.State = SandboxStateIdle
	s.record.IdleUntilEpoch = 0
	return nil
}

func (s *fakeSandboxStore) AbortStop(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.record.State == SandboxStateStopping {
		s.record.State = SandboxStateIdle
	}
	return nil
}

func (s *fakeSandboxStore) seedLease(leaseID string, expires int64, state SandboxState, activeCount int64, idleUntil int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.leases[leaseID] = expires
	s.record.State = state
	s.record.ActiveCount = activeCount
	s.record.IdleUntilEpoch = idleUntil
	s.record.Leases = copyLeaseMap(s.leases)
}

type fakeSandboxEC2 struct {
	mu       sync.Mutex
	state    string
	events   *[]string
	starts   int
	stops    int
	errStart error
	errStop  error
	errWait  error
}

func (e *fakeSandboxEC2) DescribeInstance(context.Context, string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.state, nil
}

func (e *fakeSandboxEC2) StartInstance(context.Context, string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.errStart != nil {
		return e.errStart
	}
	e.starts++
	e.state = "pending"
	if e.events != nil {
		*e.events = append(*e.events, "start")
	}
	return nil
}

func (e *fakeSandboxEC2) WaitInstanceRunning(context.Context, string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.errWait != nil {
		return e.errWait
	}
	e.state = "running"
	if e.events != nil {
		*e.events = append(*e.events, "ec2-running")
	}
	return nil
}

func (e *fakeSandboxEC2) StopInstance(context.Context, string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.errStop != nil {
		return e.errStop
	}
	e.stops++
	e.state = "stopping"
	if e.events != nil {
		*e.events = append(*e.events, "stop")
	}
	return nil
}

func (e *fakeSandboxEC2) WaitInstanceStopped(context.Context, string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.errWait != nil {
		return e.errWait
	}
	e.state = "stopped"
	return nil
}

type fakeSandboxHealth struct {
	err    error
	events *[]string
}

func newTestSandboxLifecycle(
	store SandboxLifecycleStore,
	instance SandboxInstanceClient,
	broker SandboxBrokerHealthChecker,
	now func() time.Time,
) *SandboxLifecycle {
	lifecycle, err := NewSandboxLifecycle(store, instance, broker, "i-0123456789abcdef0", time.Minute)
	if err != nil {
		panic(err)
	}
	lifecycle.now = now
	return lifecycle
}

func (h *fakeSandboxHealth) CheckAvailable(context.Context) error {
	if h.events != nil {
		*h.events = append(*h.events, "broker-healthy")
	}
	return h.err
}

func TestSandboxLifecycle_StartsBeforeEvolution(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	events := []string{}
	store := newFakeSandboxStore()
	ec2 := &fakeSandboxEC2{state: "stopped", events: &events}
	broker := &fakeSandboxHealth{events: &events}
	lifecycle := newTestSandboxLifecycle(store, ec2, broker, func() time.Time { return now })

	release, err := lifecycle.Acquire(context.Background())
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	events = append(events, "evolution")
	if err := release(); err != nil {
		t.Fatalf("release: %v", err)
	}

	want := []string{"start", "ec2-running", "broker-healthy", "evolution"}
	if fmt.Sprint(events) != fmt.Sprint(want) {
		t.Fatalf("event order = %v, want %v", events, want)
	}
}

func TestSandboxLifecycle_FailsWhenReadinessTimesOut(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	wantErr := context.DeadlineExceeded
	store := newFakeSandboxStore()
	ec2 := &fakeSandboxEC2{state: "stopped", errWait: wantErr}
	lifecycle := newTestSandboxLifecycle(store, ec2, &fakeSandboxHealth{}, func() time.Time { return now })
	lifecycle.readyTimeout = 20 * time.Millisecond

	_, err := lifecycle.Acquire(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Acquire() error = %v, want readiness timeout %v", err, wantErr)
	}
	if ec2.starts != 1 {
		t.Fatalf("start calls = %d, want 1", ec2.starts)
	}
}

func TestSandboxLifecycle_DoesNotStopWithActiveLease(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	store := newFakeSandboxStore()
	ec2 := &fakeSandboxEC2{state: "stopped"}
	lifecycle := newTestSandboxLifecycle(store, ec2, &fakeSandboxHealth{}, func() time.Time { return now })
	release, err := lifecycle.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if err := lifecycle.Reconcile(context.Background(), now.Add(24*time.Hour)); err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if ec2.stops != 0 {
		t.Fatalf("stop calls = %d while lease is active, want 0", ec2.stops)
	}
	if err := release(); err != nil {
		t.Fatal(err)
	}
}

func TestSandboxLifecycle_StopsAfterTenIdleMinutes(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	store := newFakeSandboxStore()
	ec2 := &fakeSandboxEC2{state: "stopped"}
	lifecycle := newTestSandboxLifecycle(store, ec2, &fakeSandboxHealth{}, func() time.Time { return now })
	release, err := lifecycle.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := release(); err != nil {
		t.Fatal(err)
	}

	before := now.Add(sandboxIdleDuration - time.Second)
	if err := lifecycle.Reconcile(context.Background(), before); err != nil {
		t.Fatal(err)
	}
	if ec2.stops != 0 {
		t.Fatalf("stopped before idle deadline; calls=%d", ec2.stops)
	}
	if err := lifecycle.Reconcile(context.Background(), now.Add(sandboxIdleDuration)); err != nil {
		t.Fatal(err)
	}
	if ec2.stops != 1 {
		t.Fatalf("stop calls = %d at idle deadline, want 1", ec2.stops)
	}
}

func TestSandboxLifecycle_NewLeaseCancelsIdleStop(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	current := now
	store := newFakeSandboxStore()
	ec2 := &fakeSandboxEC2{state: "stopped"}
	lifecycle := newTestSandboxLifecycle(store, ec2, &fakeSandboxHealth{}, func() time.Time { return current })
	releaseFirst, err := lifecycle.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := releaseFirst(); err != nil {
		t.Fatal(err)
	}

	current = now.Add(5 * time.Minute)
	releaseSecond, err := lifecycle.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := lifecycle.Reconcile(context.Background(), now.Add(20*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if ec2.stops != 0 {
		t.Fatalf("stop calls = %d with a new active lease, want 0", ec2.stops)
	}
	if err := releaseSecond(); err != nil {
		t.Fatal(err)
	}
}

func TestSandboxWatchdog_StopsAfterStaleLease(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	store := newFakeSandboxStore()
	store.seedLease("crashed-process-lease", now.Unix(), SandboxStateRunning, 1, 0)
	ec2 := &fakeSandboxEC2{state: "running"}
	lifecycle := newTestSandboxLifecycle(store, ec2, &fakeSandboxHealth{}, func() time.Time { return now })

	if err := lifecycle.Reconcile(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	record, err := store.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if record.ActiveCount != 0 || record.State != SandboxStateIdle {
		t.Fatalf("stale lease record = %+v, want idle with zero active leases", record)
	}
	if ec2.stops != 0 {
		t.Fatal("watchdog stopped before the stale lease's ten-minute idle grace elapsed")
	}

	if err := lifecycle.Reconcile(context.Background(), now.Add(sandboxIdleDuration)); err != nil {
		t.Fatal(err)
	}
	if ec2.stops != 1 {
		t.Fatalf("watchdog stop calls = %d, want 1 after stale lease grace", ec2.stops)
	}
}

func TestSandboxLifecycle_ReportsAWSStartErrors(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	wantErr := errors.New("EC2 StartInstances denied")
	store := newFakeSandboxStore()
	ec2 := &fakeSandboxEC2{state: "stopped", errStart: wantErr}
	lifecycle := newTestSandboxLifecycle(store, ec2, &fakeSandboxHealth{}, func() time.Time { return now })

	_, err := lifecycle.Acquire(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Acquire() error = %v, want explicit AWS error %v", err, wantErr)
	}
}

func TestSandboxLifecycle_ConcurrentLeasePreventsStopClaim(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	store := newFakeSandboxStore()
	ec2 := &fakeSandboxEC2{state: "stopped"}
	lifecycle := newTestSandboxLifecycle(store, ec2, &fakeSandboxHealth{}, func() time.Time { return now })

	release, err := lifecycle.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = lifecycle.Reconcile(context.Background(), now.Add(24*time.Hour))
	}()
	wg.Wait()
	if ec2.stops != 0 {
		t.Fatalf("stop calls = %d during active lease, want 0", ec2.stops)
	}
	if err := release(); err != nil {
		t.Fatal(err)
	}
}

func TestSandboxLifecycleRecord_UsesSharedContract(t *testing.T) {
	record := SandboxLifecycleRecord{
		ID:             sandboxLifecycleRecordID,
		State:          SandboxStateIdle,
		ActiveCount:    0,
		IdleUntilEpoch: 0,
		Leases:         map[string]int64{"opaque-id": 1_800_000_900},
	}
	if record.ID != "sandbox-lifecycle" {
		t.Fatalf("record id = %q, want sandbox-lifecycle", record.ID)
	}
	if record.Leases["opaque-id"] != 1_800_000_900 {
		t.Fatal("lease expiry must be persisted for watchdog stale-lease recovery")
	}
}

type validationDynamoLifecycleClient struct{}

func (validationDynamoLifecycleClient) GetItem(context.Context, *dynamodb.GetItemInput, ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return nil, nil
}

func (validationDynamoLifecycleClient) PutItem(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return nil, nil
}

func (validationDynamoLifecycleClient) UpdateItem(context.Context, *dynamodb.UpdateItemInput, ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return nil, nil
}

func TestNewDynamoSandboxLifecycleStore_AcceptsValidTableNames(t *testing.T) {
	for _, tableName := range []string{
		"sandbox-lifecycle",
		"ucs503-analysis-v1-123456789abc",
		strings.Repeat("a", 255),
	} {
		t.Run(tableName, func(t *testing.T) {
			store, err := NewDynamoSandboxLifecycleStore(validationDynamoLifecycleClient{}, tableName)
			if err != nil {
				t.Fatalf("NewDynamoSandboxLifecycleStore(%q) error = %v", tableName, err)
			}
			if store.tableName != tableName {
				t.Fatalf("table name = %q, want %q", store.tableName, tableName)
			}
		})
	}
}

func TestNewDynamoSandboxLifecycleStore_RejectsInvalidTableNames(t *testing.T) {
	for _, tableName := range []string{
		"ab",
		strings.Repeat("a", 256),
		"bad/name",
		"has space",
		"bad_表",
	} {
		t.Run(tableName, func(t *testing.T) {
			if _, err := NewDynamoSandboxLifecycleStore(validationDynamoLifecycleClient{}, tableName); err == nil {
				t.Fatalf("NewDynamoSandboxLifecycleStore(%q) error = nil, want invalid table name error", tableName)
			}
		})
	}
}
