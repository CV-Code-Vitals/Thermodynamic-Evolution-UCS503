package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func TestDynamoAccountStoreAtomicTicketsAndRepositoryLimit(t *testing.T) {
	endpoint := os.Getenv("DYNAMODB_ACCOUNT_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("Set DYNAMODB_ACCOUNT_TEST_ENDPOINT to a local DynamoDB test server")
	}
	if endpoint != "http://127.0.0.1:18192" {
		t.Fatal("account integration tests only run against the disposable loopback DynamoDB server")
	}
	client := dynamodb.NewFromConfig(aws.Config{Region: "us-east-1", Credentials: aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
		return aws.Credentials{AccessKeyID: "localtest", SecretAccessKey: "localtest"}, nil
	})}, func(options *dynamodb.Options) { options.BaseEndpoint = aws.String(endpoint) })
	table := fmt.Sprintf("accounts-test-%d", time.Now().UnixNano())
	_, err := client.CreateTable(context.Background(), &dynamodb.CreateTableInput{TableName: aws.String(table),
		AttributeDefinitions: []types.AttributeDefinition{{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS}},
		KeySchema:            []types.KeySchemaElement{{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash}}, BillingMode: types.BillingModePayPerRequest})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{TableName: aws.String(table)})
	})
	exerciseAccountStore(t, &dynamoAccountStore{client: client, table: table})
}
