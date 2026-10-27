package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var errAccountRecordMissing = errors.New("account record not found")
var errSavedRepositoryLimit = errors.New("saved repository limit reached")

const savedRepositoryLimit = 50

type GitHubUser struct {
	ID    string `json:"id" dynamodbav:"id"`
	Login string `json:"login" dynamodbav:"login"`
}

type SavedRepository struct {
	ID      string    `json:"id" dynamodbav:"id"`
	URL     string    `json:"url" dynamodbav:"url"`
	Name    string    `json:"name" dynamodbav:"name"`
	SavedAt time.Time `json:"saved_at" dynamodbav:"savedAt"`
}

type accountTransient struct {
	ID               string     `json:"id" dynamodbav:"id"`
	User             GitHubUser `json:"user" dynamodbav:"user"`
	Verifier         string     `json:"verifier,omitempty" dynamodbav:"verifier,omitempty"`
	BrowserChallenge string     `json:"browserChallenge,omitempty" dynamodbav:"browserChallenge,omitempty"`
	ExpiresAt        int64      `json:"expiresAt" dynamodbav:"expiresAt"`
}

type accountStore interface {
	PutTransient(context.Context, string, accountTransient) error
	GetTransient(context.Context, string) (accountTransient, error)
	TakeTransient(context.Context, string) (accountTransient, error)
	ListRepositories(context.Context, string) ([]SavedRepository, error)
	SaveRepository(context.Context, string, SavedRepository) error
	DeleteRepository(context.Context, string, string) error
}

func sortedRepositories(repositories map[string]SavedRepository) []SavedRepository {
	result := make([]SavedRepository, 0, len(repositories))
	for _, repository := range repositories {
		result = append(result, repository)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].SavedAt.Equal(result[j].SavedAt) {
			return result[i].Name < result[j].Name
		}
		return result[i].SavedAt.After(result[j].SavedAt)
	})
	return result
}

type localAccountData struct {
	Transients   map[string]accountTransient           `json:"transients"`
	Repositories map[string]map[string]SavedRepository `json:"repositories"`
}

type localAccountStore struct {
	path string
	mu   sync.Mutex
}

func (s *localAccountStore) access(ctx context.Context, write bool, action func(*localAccountData) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	data := localAccountData{Transients: map[string]accountTransient{}, Repositories: map[string]map[string]SavedRepository{}}
	file, err := os.Open(s.path)
	if err == nil {
		err = json.NewDecoder(io.LimitReader(file, 8<<20)).Decode(&data)
		_ = file.Close()
		if err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if data.Transients == nil || data.Repositories == nil {
		return errors.New("invalid account data")
	}
	if write {
		for key, record := range data.Transients {
			if record.ExpiresAt <= time.Now().Unix() {
				delete(data.Transients, key)
			}
		}
	}
	if err := action(&data); err != nil {
		return err
	}
	if !write {
		return nil
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if len(encoded) >= 8<<20 {
		return errors.New("local account storage capacity reached")
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(s.path), ".accounts-*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	if _, err = temporary.Write(encoded); err != nil {
		_ = temporary.Close()
		return err
	}
	if err = temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err = temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), s.path)
}

func (s *localAccountStore) PutTransient(ctx context.Context, key string, record accountTransient) error {
	return s.access(ctx, true, func(data *localAccountData) error {
		if _, exists := data.Transients[key]; exists {
			return errors.New("account transient already exists")
		}
		if len(data.Transients) >= 1000 {
			return errors.New("account sign-in capacity reached")
		}
		data.Transients[key] = record
		return nil
	})
}

func (s *localAccountStore) readTransient(ctx context.Context, key string, consume bool) (accountTransient, error) {
	var record accountTransient
	err := s.access(ctx, consume, func(data *localAccountData) error {
		var ok bool
		record, ok = data.Transients[key]
		if !ok {
			return errAccountRecordMissing
		}
		if consume {
			delete(data.Transients, key)
		}
		return nil
	})
	return record, err
}
func (s *localAccountStore) GetTransient(ctx context.Context, key string) (accountTransient, error) {
	return s.readTransient(ctx, key, false)
}
func (s *localAccountStore) TakeTransient(ctx context.Context, key string) (accountTransient, error) {
	return s.readTransient(ctx, key, true)
}
func (s *localAccountStore) ListRepositories(ctx context.Context, user string) ([]SavedRepository, error) {
	var result []SavedRepository
	err := s.access(ctx, false, func(data *localAccountData) error { result = sortedRepositories(data.Repositories[user]); return nil })
	return result, err
}
func (s *localAccountStore) SaveRepository(ctx context.Context, user string, repository SavedRepository) error {
	return s.access(ctx, true, func(data *localAccountData) error {
		if data.Repositories[user] == nil {
			data.Repositories[user] = map[string]SavedRepository{}
		}
		if _, exists := data.Repositories[user][repository.ID]; !exists && len(data.Repositories[user]) >= savedRepositoryLimit {
			return errSavedRepositoryLimit
		}
		data.Repositories[user][repository.ID] = repository
		return nil
	})
}
func (s *localAccountStore) DeleteRepository(ctx context.Context, user, id string) error {
	return s.access(ctx, true, func(data *localAccountData) error { delete(data.Repositories[user], id); return nil })
}

type dynamoAccountStore struct {
	client dynamoDBAPI
	table  string
}

func (s *dynamoAccountStore) key(id string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: id}}
}
func (s *dynamoAccountStore) PutTransient(ctx context.Context, key string, record accountTransient) error {
	record.ID = key
	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return err
	}
	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(s.table), Item: item, ConditionExpression: aws.String("attribute_not_exists(id)")})
	return err
}
func decodeAccountTransient(item map[string]types.AttributeValue) (accountTransient, error) {
	if len(item) == 0 {
		return accountTransient{}, errAccountRecordMissing
	}
	var record accountTransient
	err := attributevalue.UnmarshalMap(item, &record)
	return record, err
}
func (s *dynamoAccountStore) GetTransient(ctx context.Context, key string) (accountTransient, error) {
	output, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{TableName: aws.String(s.table), Key: s.key(key), ConsistentRead: aws.Bool(true)})
	if err != nil {
		return accountTransient{}, err
	}
	if output == nil {
		return accountTransient{}, errors.New("empty account storage response")
	}
	return decodeAccountTransient(output.Item)
}
func (s *dynamoAccountStore) TakeTransient(ctx context.Context, key string) (accountTransient, error) {
	output, err := s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{TableName: aws.String(s.table), Key: s.key(key), ReturnValues: types.ReturnValueAllOld, ConditionExpression: aws.String("attribute_exists(id)")})
	var conditional *types.ConditionalCheckFailedException
	if errors.As(err, &conditional) {
		return accountTransient{}, errAccountRecordMissing
	}
	if err != nil {
		return accountTransient{}, err
	}
	if output == nil {
		return accountTransient{}, errors.New("empty account storage response")
	}
	return decodeAccountTransient(output.Attributes)
}
func (s *dynamoAccountStore) ListRepositories(ctx context.Context, user string) ([]SavedRepository, error) {
	output, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{TableName: aws.String(s.table), Key: s.key("repositories:" + user), ConsistentRead: aws.Bool(true)})
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, errors.New("empty account storage response")
	}
	var record struct {
		Repositories map[string]SavedRepository `dynamodbav:"repositories"`
	}
	if err := attributevalue.UnmarshalMap(output.Item, &record); err != nil {
		return nil, err
	}
	return sortedRepositories(record.Repositories), nil
}
func (s *dynamoAccountStore) SaveRepository(ctx context.Context, user string, repository SavedRepository) error {
	key := s.key("repositories:" + user)
	initial := s.key("repositories:" + user)
	initial["repositories"] = &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}}
	_, err := s.client.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(s.table), Item: initial, ConditionExpression: aws.String("attribute_not_exists(id)")})
	var conditional *types.ConditionalCheckFailedException
	if err != nil && !errors.As(err, &conditional) {
		return err
	}
	value, err := attributevalue.Marshal(repository)
	if err != nil {
		return err
	}
	_, err = s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.table), Key: key,
		UpdateExpression:          aws.String("SET #repositories.#repository = :repository"),
		ConditionExpression:       aws.String("attribute_exists(#repositories.#repository) OR size(#repositories) < :limit"),
		ExpressionAttributeNames:  map[string]string{"#repositories": "repositories", "#repository": repository.ID},
		ExpressionAttributeValues: map[string]types.AttributeValue{":repository": value, ":limit": &types.AttributeValueMemberN{Value: strconv.Itoa(savedRepositoryLimit)}},
	})
	if errors.As(err, &conditional) {
		return errSavedRepositoryLimit
	}
	return err
}
func (s *dynamoAccountStore) DeleteRepository(ctx context.Context, user, id string) error {
	_, err := s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.table), Key: s.key("repositories:" + user),
		UpdateExpression: aws.String("REMOVE #repositories.#repository"), ConditionExpression: aws.String("attribute_exists(id)"),
		ExpressionAttributeNames: map[string]string{"#repositories": "repositories", "#repository": id},
	})
	var conditional *types.ConditionalCheckFailedException
	if errors.As(err, &conditional) {
		return nil
	}
	return err
}
