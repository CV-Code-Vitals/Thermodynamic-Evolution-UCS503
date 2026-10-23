package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const archiveStateIndexName = "state-uploadedAt-index"

type dynamoDBAPI interface {
	PutItem(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(context.Context, *dynamodb.GetItemInput, ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	Query(context.Context, *dynamodb.QueryInput, ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	UpdateItem(context.Context, *dynamodb.UpdateItemInput, ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	DeleteItem(context.Context, *dynamodb.DeleteItemInput, ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

type s3ArchiveAPI interface {
	HeadObject(context.Context, *s3.HeadObjectInput, ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	CopyObject(context.Context, *s3.CopyObjectInput, ...func(*s3.Options)) (*s3.CopyObjectOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type s3ArchivePresignerAPI interface {
	PresignPostObject(context.Context, *s3.PutObjectInput, ...func(*s3.PresignPostOptions)) (*s3.PresignedPostRequest, error)
	PresignGetObject(context.Context, *s3.GetObjectInput, ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

type DynamoDBArchiveRepository struct {
	client    dynamoDBAPI
	tableName string
}

func NewDynamoDBArchiveRepository(client *dynamodb.Client, tableName string) *DynamoDBArchiveRepository {
	return newDynamoDBArchiveRepository(client, tableName)
}

func newDynamoDBArchiveRepository(client dynamoDBAPI, tableName string) *DynamoDBArchiveRepository {
	return &DynamoDBArchiveRepository{client: client, tableName: tableName}
}

func (r *DynamoDBArchiveRepository) CreatePending(ctx context.Context, entry ArchiveEntry) error {
	if entry.ID == "" || entry.State != "pending" || entry.ExpiresAt <= 0 ||
		!validGeneratedObjectKey(entry.PendingKey, "pending/") ||
		!validGeneratedObjectKey(entry.ObjectKey, "archives/") {
		return errors.New("pending archive record has invalid state or object keys")
	}
	item, err := attributevalue.MarshalMap(entry)
	if err != nil {
		return fmt.Errorf("marshal pending archive record: %w", err)
	}
	callCtx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()
	_, err = r.client.PutItem(callCtx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	})
	if err != nil {
		return fmt.Errorf("put pending archive record: %w", err)
	}
	return nil
}

func (r *DynamoDBArchiveRepository) Get(ctx context.Context, id string) (ArchiveEntry, error) {
	callCtx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()
	output, err := r.client.GetItem(callCtx, &dynamodb.GetItemInput{
		TableName:      aws.String(r.tableName),
		Key:            map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: id}},
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return ArchiveEntry{}, fmt.Errorf("get archive record: %w", err)
	}
	if output == nil {
		return ArchiveEntry{}, errors.New("get archive record returned an empty response")
	}
	if len(output.Item) == 0 {
		return ArchiveEntry{}, ErrArchiveNotFound
	}
	var entry ArchiveEntry
	if err := attributevalue.UnmarshalMap(output.Item, &entry); err != nil {
		return ArchiveEntry{}, fmt.Errorf("unmarshal archive record: %w", err)
	}
	return entry, nil
}

func (r *DynamoDBArchiveRepository) ListAvailable(ctx context.Context) ([]ArchiveEntry, error) {
	var entries []ArchiveEntry
	var startKey map[string]types.AttributeValue
	for {
		callCtx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
		output, err := r.client.Query(callCtx, &dynamodb.QueryInput{
			TableName:                 aws.String(r.tableName),
			IndexName:                 aws.String(archiveStateIndexName),
			KeyConditionExpression:    aws.String("#state = :available"),
			ExpressionAttributeNames:  map[string]string{"#state": "state"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":available": &types.AttributeValueMemberS{Value: "available"}},
			ScanIndexForward:          aws.Bool(false),
			ExclusiveStartKey:         startKey,
		})
		cancel()
		if err != nil {
			return nil, fmt.Errorf("query available archives: %w", err)
		}
		if output == nil {
			return nil, errors.New("query available archives returned an empty response")
		}
		var page []ArchiveEntry
		if err := attributevalue.UnmarshalListOfMaps(output.Items, &page); err != nil {
			return nil, fmt.Errorf("unmarshal available archive records: %w", err)
		}
		entries = append(entries, page...)
		if len(output.LastEvaluatedKey) == 0 {
			break
		}
		startKey = output.LastEvaluatedKey
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].UploadedAt == entries[j].UploadedAt {
			return entries[i].ID > entries[j].ID
		}
		return entries[i].UploadedAt > entries[j].UploadedAt
	})
	return entries, nil
}

func (r *DynamoDBArchiveRepository) SetStatus(ctx context.Context, id, state string) error {
	if state != "pending" && state != "available" && state != "deleting" {
		return fmt.Errorf("unsupported archive state %q", state)
	}
	updateExpression := "SET #state = :state"
	if state == "available" {
		updateExpression += " REMOVE expiresAt"
	}
	callCtx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()
	_, err := r.client.UpdateItem(callCtx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.tableName),
		Key:                       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: id}},
		UpdateExpression:          aws.String(updateExpression),
		ConditionExpression:       aws.String("attribute_exists(id)"),
		ExpressionAttributeNames:  map[string]string{"#state": "state"},
		ExpressionAttributeValues: map[string]types.AttributeValue{":state": &types.AttributeValueMemberS{Value: state}},
	})
	if err != nil {
		return fmt.Errorf("update archive state: %w", err)
	}
	return nil
}

func (r *DynamoDBArchiveRepository) Delete(ctx context.Context, id string) error {
	callCtx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()
	_, err := r.client.DeleteItem(callCtx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key:       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: id}},
	})
	if err != nil {
		return fmt.Errorf("delete archive record: %w", err)
	}
	return nil
}

type S3ObjectStore struct {
	client    s3ArchiveAPI
	presigner s3ArchivePresignerAPI
	bucket    string
}

func NewS3ObjectStore(client *s3.Client, presigner *s3.PresignClient, bucket string) *S3ObjectStore {
	return newS3ObjectStore(client, presigner, bucket)
}

func newS3ObjectStore(client s3ArchiveAPI, presigner s3ArchivePresignerAPI, bucket string) *S3ObjectStore {
	return &S3ObjectStore{client: client, presigner: presigner, bucket: bucket}
}

func (s *S3ObjectStore) CreateUpload(ctx context.Context, pendingKey, contentType string, size int64, expiresIn time.Duration) (PresignedUpload, error) {
	if !validGeneratedObjectKey(pendingKey, "pending/") {
		return PresignedUpload{}, errors.New("upload key must be a generated pending/ key")
	}
	if size < 1 || size > maxArchiveSize {
		return PresignedUpload{}, ErrInvalidUpload
	}
	if err := validateUploadMetadata("upload"+extensionForContentType(contentType), contentType, size); err != nil {
		return PresignedUpload{}, err
	}
	if expiresIn <= 0 || expiresIn > archiveURLLifetime {
		return PresignedUpload{}, errors.New("upload URL expiry must be between 0 and 5 minutes")
	}
	expiresAt := time.Now().UTC().Add(expiresIn)
	callCtx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()
	result, err := s.presigner.PresignPostObject(callCtx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(pendingKey),
		ContentType: aws.String(contentType),
	}, func(options *s3.PresignPostOptions) {
		options.Expires = expiresIn
		options.Conditions = append(options.Conditions,
			map[string]string{"bucket": s.bucket},
			[]any{"eq", "$key", pendingKey},
			[]any{"eq", "$Content-Type", contentType},
			[]any{"content-length-range", int64(1), maxArchiveSize},
		)
	})
	if err != nil {
		return PresignedUpload{}, fmt.Errorf("presign S3 upload: %w", err)
	}
	if result == nil {
		return PresignedUpload{}, errors.New("presign S3 upload returned an empty response")
	}
	fields := make(map[string]string, len(result.Values))
	for name, value := range result.Values {
		fields[name] = value
	}
	fields["key"] = pendingKey
	fields["Content-Type"] = contentType
	return PresignedUpload{
		URL:       result.URL,
		Fields:    fields,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *S3ObjectStore) VerifyUpload(ctx context.Context, pendingKey, originalName, contentType string, expectedSize int64) error {
	if !validGeneratedObjectKey(pendingKey, "pending/") {
		return errors.New("upload key must be a generated pending/ key")
	}
	if err := validateUploadMetadata(originalName, contentType, expectedSize); err != nil {
		return err
	}
	callCtx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()
	head, err := s.client.HeadObject(callCtx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(pendingKey),
	})
	if err != nil {
		return fmt.Errorf("head staged S3 object: %w", err)
	}
	if head == nil {
		return errors.New("head staged S3 object returned an empty response")
	}
	if head.ContentLength == nil || *head.ContentLength != expectedSize ||
		head.ContentType == nil || *head.ContentType != contentType {
		return ErrInvalidUpload
	}
	object, err := s.client.GetObject(callCtx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(pendingKey),
	})
	if err != nil {
		return fmt.Errorf("read staged S3 object: %w", err)
	}
	if object == nil || object.Body == nil {
		return errors.New("read staged S3 object returned an empty body")
	}
	defer object.Body.Close()
	signature, err := io.ReadAll(io.LimitReader(object.Body, maxArchiveSize+1))
	if err != nil {
		return fmt.Errorf("read staged S3 object: %w", err)
	}
	if int64(len(signature)) != expectedSize {
		return ErrInvalidUpload
	}
	if err := validateUpload(originalName, contentType, expectedSize, signature); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidUpload, err)
	}
	return nil
}

func (s *S3ObjectStore) PromoteUpload(ctx context.Context, pendingKey, archiveKey string) error {
	if !validGeneratedObjectKey(pendingKey, "pending/") || !validGeneratedObjectKey(archiveKey, "archives/") {
		return errors.New("promotion requires generated pending/ and archives/ keys")
	}
	callCtx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()
	_, err := s.client.CopyObject(callCtx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		CopySource: aws.String(escapeCopySource(s.bucket + "/" + pendingKey)),
		Key:        aws.String(archiveKey),
	})
	if err != nil {
		return fmt.Errorf("copy staged S3 object to archive: %w", err)
	}
	return nil
}

func (s *S3ObjectStore) PresignDownload(ctx context.Context, archiveKey string, expiresIn time.Duration) (string, error) {
	if !validGeneratedObjectKey(archiveKey, "archives/") {
		return "", errors.New("download key must be a generated archives/ key")
	}
	if expiresIn <= 0 || expiresIn > archiveURLLifetime {
		return "", errors.New("download URL expiry must be between 0 and 5 minutes")
	}
	callCtx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()
	result, err := s.presigner.PresignGetObject(callCtx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(archiveKey),
	}, func(options *s3.PresignOptions) {
		options.Expires = expiresIn
	})
	if err != nil {
		return "", fmt.Errorf("presign S3 download: %w", err)
	}
	if result == nil {
		return "", errors.New("presign S3 download returned an empty response")
	}
	return result.URL, nil
}

func (s *S3ObjectStore) DeleteObject(ctx context.Context, key string) error {
	if !validGeneratedObjectKey(key, "pending/") && !validGeneratedObjectKey(key, "archives/") {
		return errors.New("delete key must be under pending/ or archives/")
	}
	callCtx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()
	_, err := s.client.DeleteObject(callCtx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete S3 object: %w", err)
	}
	return nil
}

func validGeneratedObjectKey(key, prefix string) bool {
	if !strings.HasPrefix(key, prefix) {
		return false
	}
	id := strings.TrimPrefix(key, prefix)
	if len(id) != 36 || strings.Contains(id, "/") || strings.Contains(id, "..") {
		return false
	}
	for i, char := range id {
		switch i {
		case 8, 13, 18, 23:
			if char != '-' {
				return false
			}
		default:
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
				return false
			}
		}
	}
	return true
}

func extensionForContentType(contentType string) string {
	switch contentType {
	case "application/pdf":
		return ".pdf"
	case "application/vnd.ms-powerpoint":
		return ".ppt"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return ".pptx"
	default:
		return ".unsupported"
	}
}

func escapeCopySource(source string) string {
	parts := strings.Split(source, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}

var _ ArchiveRepository = (*DynamoDBArchiveRepository)(nil)
var _ ObjectStore = (*S3ObjectStore)(nil)
