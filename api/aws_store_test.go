package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestArchiveServiceBeginUploadCreatesPrivatePendingRecord(t *testing.T) {
	repository := newFakeArchiveRepository()
	objects := newFakeArchiveObjectStore()
	service := newArchiveService(repository, objects)
	metadata := validUploadMetadata()

	upload, err := service.BeginUpload(context.Background(), metadata)
	if err != nil {
		t.Fatalf("BeginUpload returned an error: %v", err)
	}
	if len(repository.entries) != 1 {
		t.Fatalf("expected one pending record, got %d", len(repository.entries))
	}
	entry := repository.entries[objects.lastPendingKey[strings.Index(objects.lastPendingKey, "/")+1:]]
	if entry.State != "pending" || !strings.HasPrefix(entry.PendingKey, "pending/") {
		t.Fatalf("pending record has invalid state or staging key: %#v", entry)
	}
	if entry.ObjectKey == "" || !strings.HasPrefix(entry.ObjectKey, "archives/") {
		t.Fatalf("permanent archive key was not generated separately: %#v", entry)
	}
	if entry.OriginalName != metadata.OriginalName || entry.Size != metadata.Size {
		t.Fatalf("pending record did not preserve validated metadata: %#v", entry)
	}
	if objects.lastExpiresIn != 5*time.Minute {
		t.Fatalf("upload URL lifetime = %s, want 5m", objects.lastExpiresIn)
	}
	if upload.URL == "" || !upload.ExpiresAt.After(time.Now()) {
		t.Fatalf("expected a usable presigned upload result, got %#v", upload)
	}
}

func TestArchiveServiceBeginUploadRejectsInvalidMetadataBeforeStorage(t *testing.T) {
	repository := newFakeArchiveRepository()
	objects := newFakeArchiveObjectStore()
	service := newArchiveService(repository, objects)

	metadata := validUploadMetadata()
	metadata.ContentType = "application/zip"
	if _, err := service.BeginUpload(context.Background(), metadata); err == nil {
		t.Fatal("BeginUpload accepted a mismatched content type")
	}
	if len(repository.entries) != 0 || objects.createUploadCalls != 0 {
		t.Fatal("invalid upload metadata reached storage")
	}
}

func TestArchiveServiceCompleteUploadVerifiesThenPromotesAndPublishes(t *testing.T) {
	repository := newFakeArchiveRepository()
	objects := newFakeArchiveObjectStore()
	service := newArchiveService(repository, objects)
	_, err := service.BeginUpload(context.Background(), validUploadMetadata())
	if err != nil {
		t.Fatal(err)
	}
	id := uploadIDFromFake(repository)

	entry, err := service.CompleteUpload(context.Background(), id)
	if err != nil {
		t.Fatalf("CompleteUpload returned an error: %v", err)
	}
	if entry.State != "available" || !strings.HasPrefix(entry.ObjectKey, "archives/") {
		t.Fatalf("completed entry is not available under archives/: %#v", entry)
	}
	if objects.verifyCalls != 1 || objects.promoteCalls != 1 {
		t.Fatalf("verification/promotion calls = %d/%d, want 1/1", objects.verifyCalls, objects.promoteCalls)
	}
	if objects.lastPromoteSource != entry.PendingKey || objects.lastPromoteTarget != entry.ObjectKey {
		t.Fatalf("promoted unexpected keys: %q -> %q", objects.lastPromoteSource, objects.lastPromoteTarget)
	}
	if repository.entries[id].State != "available" {
		t.Fatal("archive metadata became available only in the response, not the repository")
	}
	if !containsString(objects.deletedKeys, entry.PendingKey) {
		t.Fatal("staging object was not removed after successful promotion")
	}
}

func TestArchiveServiceCompleteUploadRejectsInvalidOrMissingStagedObjects(t *testing.T) {
	for _, scenario := range []string{"missing", "wrong size", "wrong content type", "wrong signature", "invalid OOXML"} {
		t.Run(scenario, func(t *testing.T) {
			repository := newFakeArchiveRepository()
			objects := newFakeArchiveObjectStore()
			service := newArchiveService(repository, objects)
			_, err := service.BeginUpload(context.Background(), validUploadMetadata())
			if err != nil {
				t.Fatal(err)
			}
			id := uploadIDFromFake(repository)
			objects.verifyErr = fmt.Errorf("%s: %w", scenario, ErrInvalidUpload)

			if _, err := service.CompleteUpload(context.Background(), id); err == nil {
				t.Fatal("CompleteUpload accepted an invalid staged object")
			}
			if repository.entries[id].State != "pending" || objects.promoteCalls != 0 {
				t.Fatal("invalid staged object was promoted or published")
			}
		})
	}
}

func TestArchiveServiceListReturnsNewestFirst(t *testing.T) {
	repository := newFakeArchiveRepository()
	repository.listEntries = []ArchiveEntry{
		{ID: "old", State: "available", ObjectKey: "archives/00000000-0000-4000-8000-000000000001", UploadedAt: "2026-10-20T10:00:00.000000000Z"},
		{ID: "new", State: "available", ObjectKey: "archives/00000000-0000-4000-8000-000000000002", UploadedAt: "2026-10-22T10:00:00.000000000Z"},
		{ID: "middle", State: "available", ObjectKey: "archives/00000000-0000-4000-8000-000000000003", UploadedAt: "2026-10-21T10:00:00.000000000Z"},
	}
	entries, err := newArchiveService(repository, newFakeArchiveObjectStore()).List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got := []string{entries[0].ID, entries[1].ID, entries[2].ID}; strings.Join(got, ",") != "new,middle,old" {
		t.Fatalf("list order = %v, want newest first", got)
	}
}

func TestArchiveServiceListReturnsEmptyArrayForEmptyArchive(t *testing.T) {
	entries, err := newArchiveService(newFakeArchiveRepository(), newFakeArchiveObjectStore()).List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if entries == nil {
		t.Fatal("empty archive list is nil; the API would encode it as JSON null instead of []")
	}
	if len(entries) != 0 {
		t.Fatalf("empty archive list has %d entries, want 0", len(entries))
	}
}

func TestArchiveServiceDownloadRequiresAvailableAndUsesFiveMinuteURL(t *testing.T) {
	repository := newFakeArchiveRepository()
	objects := newFakeArchiveObjectStore()
	repository.entries["ready"] = ArchiveEntry{ID: "ready", State: "available", ObjectKey: "archives/00000000-0000-4000-8000-000000000001"}
	repository.entries["pending"] = ArchiveEntry{ID: "pending", State: "pending", ObjectKey: "archives/00000000-0000-4000-8000-000000000002"}
	service := newArchiveService(repository, objects)

	result, err := service.Download(context.Background(), "ready")
	if err != nil {
		t.Fatal(err)
	}
	if objects.lastExpiresIn != 5*time.Minute || result.URL == "" || !result.ExpiresAt.After(time.Now()) {
		t.Fatalf("download URL has wrong expiry or is empty: %#v", result)
	}
	if remaining := time.Until(result.ExpiresAt); remaining < 4*time.Minute+59*time.Second || remaining > 5*time.Minute {
		t.Fatalf("download URL expiry is not 300 seconds from issue time: %s remain", remaining)
	}
	if _, err := service.Download(context.Background(), "pending"); err == nil {
		t.Fatal("Download returned a URL for a non-available record")
	}
}

func TestArchiveServiceRejectsUploadAfterPresignedPostExpires(t *testing.T) {
	repository := newFakeArchiveRepository()
	objects := newFakeArchiveObjectStore()
	service := newArchiveService(repository, objects)
	_, err := service.BeginUpload(context.Background(), validUploadMetadata())
	if err != nil {
		t.Fatal(err)
	}
	id := uploadIDFromFake(repository)
	objects.verifyErr = errors.New("S3 rejected expired POST policy")

	if _, err := service.CompleteUpload(context.Background(), id); err == nil {
		t.Fatal("expired presigned upload was accepted without a verified staged object")
	}
	if repository.entries[id].State != "pending" || objects.promoteCalls != 0 {
		t.Fatal("expired upload was promoted or made available")
	}
}

func TestArchiveServiceCompleteRetryDoesNotOverwritePermanentObject(t *testing.T) {
	repository := newFakeArchiveRepository()
	objects := newFakeArchiveObjectStore()
	service := newArchiveService(repository, objects)
	_, err := service.BeginUpload(context.Background(), validUploadMetadata())
	if err != nil {
		t.Fatal(err)
	}
	id := uploadIDFromFake(repository)
	first, err := service.CompleteUpload(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	objects.deletedKeys = nil
	if _, err := service.CompleteUpload(context.Background(), id); err != nil {
		t.Fatalf("retry of completed upload failed: %v", err)
	}
	if objects.promoteCalls != 1 || first.ObjectKey != repository.entries[id].ObjectKey {
		t.Fatalf("retry overwrote or changed permanent archive key: calls=%d", objects.promoteCalls)
	}
	if !containsString(objects.deletedKeys, first.PendingKey) {
		t.Fatal("retry did not clean up the staging object")
	}
}

func TestArchiveServiceCompleteUploadPreservesObjectAfterOperationErrors(t *testing.T) {
	for _, stage := range []string{"copy", "metadata"} {
		t.Run(stage, func(t *testing.T) {
			repository := newFakeArchiveRepository()
			objects := newFakeArchiveObjectStore()
			service := newArchiveService(repository, objects)
			_, err := service.BeginUpload(context.Background(), validUploadMetadata())
			if err != nil {
				t.Fatal(err)
			}
			id := uploadIDFromFake(repository)
			if stage == "copy" {
				objects.promoteErr = errors.New("copy failed")
			} else {
				repository.setStatusErr = errors.New("metadata update failed")
			}

			if _, err := service.CompleteUpload(context.Background(), id); err == nil {
				t.Fatal("CompleteUpload hid a failed storage operation")
			}
			entry := repository.entries[id]
			if entry.State != "pending" {
				t.Fatalf("failed finalization published state %q", entry.State)
			}
			if containsString(objects.deletedKeys, entry.ObjectKey) {
				t.Fatal("failed operation deleted the permanent object before state was reconciled")
			}
		})
	}
}

func TestArchiveServiceCompleteUploadRetriesAfterLostPromoteResponse(t *testing.T) {
	repository := newFakeArchiveRepository()
	objects := &promoteThenErrorArchiveObjectStore{fakeArchiveObjectStore: newFakeArchiveObjectStore(), failAfterPromote: true}
	service := newArchiveService(repository, objects)
	if _, err := service.BeginUpload(context.Background(), validUploadMetadata()); err != nil {
		t.Fatal(err)
	}
	id := uploadIDFromFake(repository)

	if _, err := service.CompleteUpload(context.Background(), id); err == nil {
		t.Fatal("first completion hid the simulated lost CopyObject response")
	}
	entry := repository.entries[id]
	if entry.State != "pending" {
		t.Fatalf("failed response changed archive state to %q", entry.State)
	}
	if containsString(objects.deletedKeys, entry.ObjectKey) {
		t.Fatal("lost CopyObject response caused deletion of the possibly copied object")
	}

	finalized, err := service.CompleteUpload(context.Background(), id)
	if err != nil {
		t.Fatalf("retry after lost CopyObject response failed: %v", err)
	}
	if finalized.State != "available" {
		t.Fatalf("retry returned state %q, want available", finalized.State)
	}
	if objects.promoteCalls != 2 {
		t.Fatalf("promotion attempts = %d, want initial attempt plus retry", objects.promoteCalls)
	}
}

func TestArchiveServiceCompleteUploadDoesNotDeleteObjectAfterAmbiguousAvailableCommit(t *testing.T) {
	repository := &commitThenErrorArchiveRepository{fakeArchiveRepository: newFakeArchiveRepository()}
	objects := newFakeArchiveObjectStore()
	service := newArchiveService(repository, objects)
	if _, err := service.BeginUpload(context.Background(), validUploadMetadata()); err != nil {
		t.Fatal(err)
	}
	id := uploadIDFromFake(repository.fakeArchiveRepository)

	if _, err := service.CompleteUpload(context.Background(), id); err != nil {
		t.Fatalf("CompleteUpload did not reconcile the committed available state: %v", err)
	}
	entry := repository.entries[id]
	if entry.State != "available" {
		t.Fatalf("persisted state = %q, want available", entry.State)
	}
	if containsString(objects.deletedKeys, entry.ObjectKey) {
		t.Fatal("compensation deleted an object whose metadata commit succeeded")
	}
	if !containsString(objects.deletedKeys, entry.PendingKey) {
		t.Fatal("successful reconciliation did not clean up the staging object")
	}
	promotions := objects.promoteCalls

	if _, err := service.CompleteUpload(context.Background(), id); err != nil {
		t.Fatalf("retry of reconciled completion failed: %v", err)
	}
	if objects.promoteCalls != promotions || containsString(objects.deletedKeys, entry.ObjectKey) {
		t.Fatal("retry promoted again or deleted the available archive")
	}
}

func TestArchiveServiceCompleteUploadKeepsObjectWhenCommitStateCannotBeRead(t *testing.T) {
	repository := &commitThenErrorArchiveRepository{fakeArchiveRepository: newFakeArchiveRepository(), failStateRead: true}
	objects := newFakeArchiveObjectStore()
	service := newArchiveService(repository, objects)
	if _, err := service.BeginUpload(context.Background(), validUploadMetadata()); err != nil {
		t.Fatal(err)
	}
	id := uploadIDFromFake(repository.fakeArchiveRepository)

	if _, err := service.CompleteUpload(context.Background(), id); err == nil {
		t.Fatal("CompleteUpload hid the ambiguous state read")
	}
	entry := repository.entries[id]
	if entry.State != "available" {
		t.Fatalf("simulated lost response did not commit available state: %q", entry.State)
	}
	if containsString(objects.deletedKeys, entry.ObjectKey) {
		t.Fatal("completion deleted the permanent object without confirming persisted state")
	}

	if _, err := service.CompleteUpload(context.Background(), id); err != nil {
		t.Fatalf("retry did not recover from the ambiguous state read: %v", err)
	}
	if containsString(objects.deletedKeys, entry.ObjectKey) {
		t.Fatal("retry deleted the available permanent object")
	}
}

func TestArchiveServiceCompleteSurfacesStagingCleanupFailureAndRetryCompletes(t *testing.T) {
	repository := newFakeArchiveRepository()
	objects := newFakeArchiveObjectStore()
	service := newArchiveService(repository, objects)
	_, err := service.BeginUpload(context.Background(), validUploadMetadata())
	if err != nil {
		t.Fatal(err)
	}
	id := uploadIDFromFake(repository)
	objects.deleteErrByKey["pending/"] = errors.New("stage cleanup failed")

	if _, err := service.CompleteUpload(context.Background(), id); err == nil {
		t.Fatal("CompleteUpload hid a staging cleanup failure")
	}
	if repository.entries[id].State != "available" {
		t.Fatal("staging cleanup failure should not roll back an already-published archive")
	}
	delete(objects.deleteErrByKey, "pending/")
	if _, err := service.CompleteUpload(context.Background(), id); err != nil {
		t.Fatalf("retry did not finish staged-object cleanup: %v", err)
	}
}

func TestArchiveServiceDeleteKeepsAmbiguousObjectDeletionRecoverable(t *testing.T) {
	repository := newFakeArchiveRepository()
	objects := &deleteThenErrorArchiveObjectStore{fakeArchiveObjectStore: newFakeArchiveObjectStore(), failAfterDelete: true}
	entry := ArchiveEntry{ID: "entry", State: "available", ObjectKey: "archives/00000000-0000-4000-8000-000000000003"}
	repository.entries[entry.ID] = entry
	service := newArchiveService(repository, objects)

	if err := service.Delete(context.Background(), entry.ID); err == nil {
		t.Fatal("Delete hid the lost response after S3 removed the object")
	}
	if repository.entries[entry.ID].State != "deleting" {
		t.Fatalf("ambiguous S3 deletion restored state %q, want deleting", repository.entries[entry.ID].State)
	}

	if err := service.Delete(context.Background(), entry.ID); err != nil {
		t.Fatalf("retry did not finish deletion from the recoverable state: %v", err)
	}
	if _, exists := repository.entries[entry.ID]; exists {
		t.Fatal("retry did not remove the deleted archive metadata")
	}
}

func TestArchiveServiceDeleteTreatsAlreadyDeletedRecordAsSuccess(t *testing.T) {
	repository := &deleteThenErrorArchiveRepository{fakeArchiveRepository: newFakeArchiveRepository()}
	objects := newFakeArchiveObjectStore()
	entry := ArchiveEntry{
		ID:        "00000000-0000-4000-8000-000000000005",
		State:     "available",
		ObjectKey: "archives/00000000-0000-4000-8000-000000000005",
	}
	repository.entries[entry.ID] = entry
	service := newArchiveService(repository, objects)

	if err := service.Delete(context.Background(), entry.ID); err != nil {
		t.Fatalf("delete did not reconcile the lost metadata-delete response: %v", err)
	}
	if _, exists := repository.entries[entry.ID]; exists {
		t.Fatal("lost metadata-delete response restored the deleted record")
	}
	if err := service.Delete(context.Background(), entry.ID); err != nil {
		t.Fatalf("retry for an already deleted record returned an error: %v", err)
	}
	entries, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("list after deletion failed: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("deleted record was exposed in the archive list: %#v", entries)
	}
}

func TestArchiveServiceDeleteRetriesObjectAndMetadataFailures(t *testing.T) {
	t.Run("retain deleting state after object deletion failure", func(t *testing.T) {
		repository := newFakeArchiveRepository()
		objects := newFakeArchiveObjectStore()
		entry := ArchiveEntry{ID: "entry", State: "available", ObjectKey: "archives/00000000-0000-4000-8000-000000000003"}
		repository.entries["entry"] = entry
		objects.deleteErrByKey[entry.ObjectKey] = errors.New("S3 unavailable")
		service := newArchiveService(repository, objects)
		if err := service.Delete(context.Background(), "entry"); err == nil {
			t.Fatal("Delete hid the S3 failure")
		}
		if repository.entries["entry"].State != "deleting" {
			t.Fatalf("failed object deletion exposed state %q, want deleting", repository.entries["entry"].State)
		}
		delete(objects.deleteErrByKey, entry.ObjectKey)
		if err := service.Delete(context.Background(), "entry"); err != nil {
			t.Fatalf("retry after object deletion failure failed: %v", err)
		}
	})

	t.Run("retry after metadata deletion failure", func(t *testing.T) {
		repository := newFakeArchiveRepository()
		objects := newFakeArchiveObjectStore()
		repository.entries["entry"] = ArchiveEntry{ID: "entry", State: "available", ObjectKey: "archives/00000000-0000-4000-8000-000000000004"}
		repository.deleteErr = errors.New("DynamoDB unavailable")
		service := newArchiveService(repository, objects)
		if err := service.Delete(context.Background(), "entry"); err == nil {
			t.Fatal("Delete hid the DynamoDB failure")
		}
		if repository.entries["entry"].State != "deleting" {
			t.Fatal("metadata failure should leave a retryable deleting state")
		}
		repository.deleteErr = nil
		if err := service.Delete(context.Background(), "entry"); err != nil {
			t.Fatalf("retry did not finish deletion: %v", err)
		}
		if _, exists := repository.entries["entry"]; exists {
			t.Fatal("retry did not remove the metadata record")
		}
	})
}

type commitThenErrorArchiveRepository struct {
	*fakeArchiveRepository
	failStateRead bool
	failNextGet   bool
}

type deleteThenErrorArchiveRepository struct {
	*fakeArchiveRepository
	failAfterDelete bool
}

func (r *deleteThenErrorArchiveRepository) Delete(ctx context.Context, id string) error {
	if err := r.fakeArchiveRepository.Delete(ctx, id); err != nil {
		return err
	}
	if !r.failAfterDelete {
		r.failAfterDelete = true
		return errors.New("simulated lost metadata-delete response")
	}
	return nil
}

func (r *commitThenErrorArchiveRepository) Get(ctx context.Context, id string) (ArchiveEntry, error) {
	if r.failNextGet {
		r.failNextGet = false
		return ArchiveEntry{}, errors.New("simulated GetItem response loss")
	}
	return r.fakeArchiveRepository.Get(ctx, id)
}

func (r *commitThenErrorArchiveRepository) SetStatus(_ context.Context, id, state string) error {
	entry, exists := r.entries[id]
	if !exists {
		return ErrArchiveNotFound
	}
	entry.State = state
	r.entries[id] = entry
	if state == "available" {
		r.failNextGet = r.failStateRead
		return errors.New("simulated UpdateItem response loss after commit")
	}
	return nil
}

type deleteThenErrorArchiveObjectStore struct {
	*fakeArchiveObjectStore
	failAfterDelete bool
}

type promoteThenErrorArchiveObjectStore struct {
	*fakeArchiveObjectStore
	failAfterPromote bool
}

func (s *promoteThenErrorArchiveObjectStore) PromoteUpload(_ context.Context, pendingKey, archiveKey string) error {
	s.promoteCalls++
	s.lastPromoteSource, s.lastPromoteTarget = pendingKey, archiveKey
	if s.failAfterPromote {
		s.failAfterPromote = false
		return errors.New("simulated CopyObject response loss after copy")
	}
	return nil
}

func (s *deleteThenErrorArchiveObjectStore) DeleteObject(_ context.Context, key string) error {
	s.deletedKeys = append(s.deletedKeys, key)
	if s.failAfterDelete {
		s.failAfterDelete = false
		return errors.New("simulated DeleteObject response loss after deletion")
	}
	return nil
}

func TestArchiveServiceSurfacesRepositoryAndObjectStoreFailures(t *testing.T) {
	t.Run("create pending record", func(t *testing.T) {
		repository := newFakeArchiveRepository()
		repository.createErr = errors.New("DynamoDB put failed")
		objects := newFakeArchiveObjectStore()
		if _, err := newArchiveService(repository, objects).BeginUpload(context.Background(), validUploadMetadata()); err == nil {
			t.Fatal("BeginUpload hid a pending-record failure")
		}
		if objects.createUploadCalls != 0 {
			t.Fatal("presigning happened after pending-record failure")
		}
	})

	t.Run("presign failure compensates pending record", func(t *testing.T) {
		repository := newFakeArchiveRepository()
		objects := newFakeArchiveObjectStore()
		objects.createUploadErr = errors.New("S3 presign failed")
		if _, err := newArchiveService(repository, objects).BeginUpload(context.Background(), validUploadMetadata()); err == nil {
			t.Fatal("BeginUpload hid a presign failure")
		}
		if len(repository.entries) != 0 {
			t.Fatal("presign failure did not compensate the pending record")
		}
	})

	t.Run("list", func(t *testing.T) {
		repository := newFakeArchiveRepository()
		repository.listErr = errors.New("DynamoDB query failed")
		if _, err := newArchiveService(repository, newFakeArchiveObjectStore()).List(context.Background()); err == nil {
			t.Fatal("List hid the repository failure")
		}
	})

	t.Run("get before finalize", func(t *testing.T) {
		repository := newFakeArchiveRepository()
		repository.getErr = errors.New("DynamoDB get failed")
		if _, err := newArchiveService(repository, newFakeArchiveObjectStore()).CompleteUpload(context.Background(), "id"); err == nil {
			t.Fatal("CompleteUpload hid the repository read failure")
		}
	})

	t.Run("verify", func(t *testing.T) {
		repository := newFakeArchiveRepository()
		objects := newFakeArchiveObjectStore()
		service := newArchiveService(repository, objects)
		_, _ = service.BeginUpload(context.Background(), validUploadMetadata())
		objects.verifyErr = errors.New("S3 verification failed")
		if _, err := service.CompleteUpload(context.Background(), uploadIDFromFake(repository)); err == nil {
			t.Fatal("CompleteUpload hid the S3 verification failure")
		}
	})

	t.Run("download presign", func(t *testing.T) {
		repository := newFakeArchiveRepository()
		repository.entries["ready"] = ArchiveEntry{ID: "ready", State: "available", ObjectKey: "archives/00000000-0000-4000-8000-000000000005"}
		objects := newFakeArchiveObjectStore()
		objects.presignDownloadErr = errors.New("S3 presign failed")
		if _, err := newArchiveService(repository, objects).Download(context.Background(), "ready"); err == nil {
			t.Fatal("Download hid the S3 presign failure")
		}
	})
}

func TestDynamoDBArchiveRepositoryWritesPendingRecordWithTTL(t *testing.T) {
	client := &fakeDynamoDBAPI{}
	repository := newDynamoDBArchiveRepository(client, "archive-table")
	entry := ArchiveEntry{
		ID:         "id",
		State:      "pending",
		PendingKey: "pending/12345678-1234-4123-8123-123456789abc",
		ObjectKey:  "archives/12345678-1234-4123-8123-123456789abc",
		ExpiresAt:  time.Now().Add(48 * time.Hour).Unix(),
	}
	if err := repository.CreatePending(context.Background(), entry); err != nil {
		t.Fatal(err)
	}
	if client.putInput == nil || aws.ToString(client.putInput.TableName) != "archive-table" {
		t.Fatal("CreatePending did not issue PutItem against the configured table")
	}
	if aws.ToString(client.putInput.ConditionExpression) != "attribute_not_exists(id)" {
		t.Fatalf("PutItem condition = %q, want unique ID guard", aws.ToString(client.putInput.ConditionExpression))
	}
	if got := client.putInput.Item["expiresAt"].(*types.AttributeValueMemberN).Value; got == "" {
		t.Fatal("pending record did not store its TTL")
	}
	if _, ok := client.putInput.Item["objectKey"]; !ok {
		t.Fatal("pending record did not store its server-generated permanent key")
	}
}

func TestDynamoDBArchiveRepositorySurfacesEveryOperationFailure(t *testing.T) {
	client := &fakeDynamoDBAPI{putErr: errors.New("put"), getErr: errors.New("get"), queryErr: errors.New("query"), updateErr: errors.New("update"), deleteErr: errors.New("delete")}
	repository := newDynamoDBArchiveRepository(client, "archive-table")
	entry := ArchiveEntry{ID: "id", State: "pending", PendingKey: "pending/12345678-1234-4123-8123-123456789abc", ObjectKey: "archives/12345678-1234-4123-8123-123456789abc", ExpiresAt: 1}
	if err := repository.CreatePending(context.Background(), entry); err == nil {
		t.Fatal("CreatePending hid a PutItem failure")
	}
	if _, err := repository.Get(context.Background(), "id"); err == nil {
		t.Fatal("Get hid a GetItem failure")
	}
	if _, err := repository.ListAvailable(context.Background()); err == nil {
		t.Fatal("ListAvailable hid a Query failure")
	}
	if err := repository.SetStatus(context.Background(), "id", "available"); err == nil {
		t.Fatal("SetStatus hid an UpdateItem failure")
	}
	if err := repository.Delete(context.Background(), "id"); err == nil {
		t.Fatal("Delete hid a DeleteItem failure")
	}
}

func TestDynamoDBArchiveRepositoryReturnsNotFoundAndPagesNewestFirst(t *testing.T) {
	t.Run("missing record", func(t *testing.T) {
		repository := newDynamoDBArchiveRepository(&fakeDynamoDBAPI{}, "archive-table")
		if _, err := repository.Get(context.Background(), "missing"); !errors.Is(err, ErrArchiveNotFound) {
			t.Fatalf("Get error = %v, want ErrArchiveNotFound", err)
		}
	})

	t.Run("paginates available index descending and strips pending TTL", func(t *testing.T) {
		client := &fakeDynamoDBAPI{
			queryOutputs: []*dynamodb.QueryOutput{
				{Items: []map[string]types.AttributeValue{
					dynamoEntryAttributes(ArchiveEntry{ID: "older", State: "available", UploadedAt: "2026-10-21T10:00:00.000000000Z"}),
				}, LastEvaluatedKey: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "older"}}},
				{Items: []map[string]types.AttributeValue{
					dynamoEntryAttributes(ArchiveEntry{ID: "newer", State: "available", UploadedAt: "2026-10-22T10:00:00.000000000Z"}),
				}},
			},
		}
		repository := newDynamoDBArchiveRepository(client, "archive-table")
		entries, err := repository.ListAvailable(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 2 || entries[0].ID != "newer" || entries[1].ID != "older" {
			t.Fatalf("paginated results were not ordered newest-first: %#v", entries)
		}
		if aws.ToString(client.queryInputs[0].IndexName) != archiveStateIndexName ||
			aws.ToBool(client.queryInputs[0].ScanIndexForward) ||
			client.queryInputs[1].ExclusiveStartKey == nil {
			t.Fatal("query did not use the state/time index, descending order, and continuation key")
		}
	})

	t.Run("available state removes TTL", func(t *testing.T) {
		client := &fakeDynamoDBAPI{}
		if err := newDynamoDBArchiveRepository(client, "archive-table").SetStatus(context.Background(), "id", "available"); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(aws.ToString(client.updateInput.UpdateExpression), "REMOVE expiresAt") {
			t.Fatalf("available state must remove pending TTL: %q", aws.ToString(client.updateInput.UpdateExpression))
		}
	})
}

func TestS3ObjectStorePresignsBoundedPendingUploadAndDownload(t *testing.T) {
	presigner := &fakeS3PresignerAPI{}
	store := newS3ObjectStore(&fakeS3ArchiveAPI{}, presigner, "private-bucket")
	pendingKey := "pending/12345678-1234-4123-8123-123456789abc"
	upload, err := store.CreateUpload(context.Background(), pendingKey, "application/pdf", 8, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if upload.URL != "https://s3.invalid/upload" || upload.Fields["key"] != pendingKey || upload.Fields["Content-Type"] != "application/pdf" {
		t.Fatalf("presigned POST fields are incomplete or wrong: %#v", upload)
	}
	if aws.ToString(presigner.postInput.Key) != pendingKey || aws.ToString(presigner.postInput.Bucket) != "private-bucket" {
		t.Fatal("presigned POST did not target the unique private pending key")
	}
	if presigner.postOptions.Expires != 5*time.Minute {
		t.Fatalf("POST expiry = %s, want 5m", presigner.postOptions.Expires)
	}
	if !hasPolicyCondition(presigner.postOptions.Conditions, []any{"eq", "$key", pendingKey}) ||
		!hasPolicyCondition(presigner.postOptions.Conditions, []any{"eq", "$Content-Type", "application/pdf"}) ||
		!hasPolicyCondition(presigner.postOptions.Conditions, []any{"content-length-range", int64(1), maxArchiveSize}) {
		t.Fatalf("POST policy omitted key, content type, or size conditions: %#v", presigner.postOptions.Conditions)
	}
	if _, err := store.CreateUpload(context.Background(), "archives/"+strings.TrimPrefix(pendingKey, "pending/"), "application/pdf", 8, 5*time.Minute); err == nil {
		t.Fatal("CreateUpload allowed a browser to write under archives/")
	}
	if _, err := store.CreateUpload(context.Background(), pendingKey, "application/pdf", 8, 6*time.Minute); err == nil {
		t.Fatal("CreateUpload allowed a URL lifetime longer than five minutes")
	}

	if _, err := store.PresignDownload(context.Background(), "archives/12345678-1234-4123-8123-123456789abc", 5*time.Minute); err != nil {
		t.Fatal(err)
	}
	if presigner.getOptions.Expires != 5*time.Minute || aws.ToString(presigner.getInput.Key) != "archives/12345678-1234-4123-8123-123456789abc" {
		t.Fatal("download URL was not scoped to the permanent object for five minutes")
	}
}

func TestS3SDKGeneratesPOSTPolicyWithoutAWSCalls(t *testing.T) {
	config := aws.Config{
		Region:      "ap-south-1",
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test-access-key", "test-secret-key", "")),
	}
	client := s3.NewFromConfig(config)
	store := newS3ObjectStore(client, s3.NewPresignClient(client), "private-bucket")
	upload, err := store.CreateUpload(context.Background(), "pending/12345678-1234-4123-8123-123456789abc", "application/pdf", 8, 5*time.Minute)
	if err != nil {
		t.Fatalf("SDK failed to generate an offline presigned POST: %v", err)
	}
	if upload.URL == "" || upload.Fields["key"] == "" || upload.Fields["Content-Type"] != "application/pdf" || upload.Fields["policy"] == "" {
		t.Fatalf("SDK presigned POST omitted required fields: %#v", upload.Fields)
	}
	policyBytes, err := base64.StdEncoding.DecodeString(upload.Fields["policy"])
	if err != nil {
		t.Fatalf("presigned POST policy is not base64 JSON: %v", err)
	}
	var policy struct {
		Expiration string `json:"expiration"`
	}
	if err := json.Unmarshal(policyBytes, &policy); err != nil {
		t.Fatalf("presigned POST policy is invalid JSON: %v", err)
	}
	policyText := string(policyBytes)
	if policy.Expiration == "" || !strings.Contains(policyText, "content-length-range") ||
		!strings.Contains(policyText, "pending/12345678-1234-4123-8123-123456789abc") ||
		!strings.Contains(policyText, "application/pdf") {
		t.Fatalf("generated POST policy is missing required constraints: %s", policyBytes)
	}
	expiration, err := time.Parse(time.RFC3339, policy.Expiration)
	remaining := time.Until(expiration)
	if err != nil || remaining < 4*time.Minute+59*time.Second || remaining > 5*time.Minute+time.Second {
		t.Fatalf("generated POST policy expiry is not five minutes: %q", policy.Expiration)
	}
}

func TestS3ObjectStorePromotesOnlyBetweenControlledPrefixes(t *testing.T) {
	client := &fakeS3ArchiveAPI{}
	store := newS3ObjectStore(client, &fakeS3PresignerAPI{}, "private-bucket")
	pendingKey := "pending/12345678-1234-4123-8123-123456789abc"
	archiveKey := "archives/abcdef12-1234-4123-8123-123456789abc"
	if err := store.PromoteUpload(context.Background(), pendingKey, archiveKey); err != nil {
		t.Fatal(err)
	}
	if client.copyCalls != 1 || client.copyTarget != archiveKey || client.copySource != "private-bucket/"+pendingKey {
		t.Fatalf("CopyObject used unexpected keys: source=%q target=%q", client.copySource, client.copyTarget)
	}
	if err := store.PromoteUpload(context.Background(), archiveKey, pendingKey); err == nil {
		t.Fatal("PromoteUpload accepted reversed or uncontrolled key prefixes")
	}
}

func TestS3ObjectStoreVerifiesPDFPPTAndPPTXContent(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        []byte
	}{
		{name: "PDF", contentType: "application/pdf", body: []byte("%PDF-1.7")},
		{name: "PPT", contentType: "application/vnd.ms-powerpoint", body: minimalPPT(t)},
		{name: "PPTX", contentType: "application/vnd.openxmlformats-officedocument.presentationml.presentation", body: minimalPPTX(t)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &fakeS3ArchiveAPI{body: test.body, contentType: test.contentType}
			store := newS3ObjectStore(client, &fakeS3PresignerAPI{}, "private-bucket")
			if err := store.VerifyUpload(context.Background(), "pending/12345678-1234-4123-8123-123456789abc", "slides"+extensionForContentType(test.contentType), test.contentType, int64(len(test.body))); err != nil {
				t.Fatalf("VerifyUpload rejected valid %s: %v", test.name, err)
			}
		})
	}
}

func TestS3ObjectStoreRejectsWrongMissingOversizedAndMalformedStagedObjects(t *testing.T) {
	valid := []byte("%PDF-1.7")
	tests := []struct {
		name        string
		body        []byte
		headSize    int64
		contentType string
		getErr      error
		headErr     error
	}{
		{name: "wrong size", body: valid, headSize: int64(len(valid) + 1), contentType: "application/pdf"},
		{name: "wrong type", body: valid, headSize: int64(len(valid)), contentType: "application/zip"},
		{name: "wrong signature", body: []byte("not a PDF"), headSize: 9, contentType: "application/pdf"},
		{name: "missing object", body: nil, headSize: 0, contentType: "application/pdf", headErr: errors.New("NoSuchKey")},
		{name: "missing body", body: nil, headSize: int64(len(valid)), contentType: "application/pdf", getErr: errors.New("GetObject failed")},
		{name: "oversized", body: []byte("x"), headSize: maxArchiveSize + 1, contentType: "application/pdf"},
		{name: "arbitrary ZIP called PPTX", body: arbitraryZIP(t), headSize: int64(len(arbitraryZIP(t))), contentType: "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &fakeS3ArchiveAPI{body: test.body, contentType: test.contentType, contentLength: test.headSize, headErr: test.headErr, getErr: test.getErr}
			store := newS3ObjectStore(client, &fakeS3PresignerAPI{}, "private-bucket")
			name, contentType, size := "slides.pdf", "application/pdf", int64(len(valid))
			if strings.Contains(test.name, "PPTX") {
				name = "slides.pptx"
				contentType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
				size = int64(len(test.body))
			}
			if err := store.VerifyUpload(context.Background(), "pending/12345678-1234-4123-8123-123456789abc", name, contentType, size); err == nil {
				t.Fatal("VerifyUpload accepted invalid or missing object data")
			}
		})
	}
}

func TestS3ObjectStoreSurfacesEveryAWSOperationFailure(t *testing.T) {
	t.Run("POST presign", func(t *testing.T) {
		presigner := &fakeS3PresignerAPI{postErr: errors.New("presign post failed")}
		if _, err := newS3ObjectStore(&fakeS3ArchiveAPI{}, presigner, "private-bucket").CreateUpload(context.Background(), "pending/12345678-1234-4123-8123-123456789abc", "application/pdf", 8, 5*time.Minute); err == nil {
			t.Fatal("CreateUpload hid POST presign failure")
		}
	})

	t.Run("HEAD", func(t *testing.T) {
		client := &fakeS3ArchiveAPI{headErr: errors.New("head failed")}
		if err := newS3ObjectStore(client, &fakeS3PresignerAPI{}, "private-bucket").VerifyUpload(context.Background(), "pending/12345678-1234-4123-8123-123456789abc", "slides.pdf", "application/pdf", 8); err == nil {
			t.Fatal("VerifyUpload hid HeadObject failure")
		}
	})

	t.Run("GET", func(t *testing.T) {
		client := &fakeS3ArchiveAPI{body: []byte("%PDF-1.7"), contentType: "application/pdf", contentLength: 8, getErr: errors.New("get failed")}
		if err := newS3ObjectStore(client, &fakeS3PresignerAPI{}, "private-bucket").VerifyUpload(context.Background(), "pending/12345678-1234-4123-8123-123456789abc", "slides.pdf", "application/pdf", 8); err == nil {
			t.Fatal("VerifyUpload hid GetObject failure")
		}
	})

	t.Run("COPY", func(t *testing.T) {
		client := &fakeS3ArchiveAPI{copyErr: errors.New("copy failed")}
		if err := newS3ObjectStore(client, &fakeS3PresignerAPI{}, "private-bucket").PromoteUpload(context.Background(), "pending/12345678-1234-4123-8123-123456789abc", "archives/12345678-1234-4123-8123-123456789abc"); err == nil {
			t.Fatal("PromoteUpload hid CopyObject failure")
		}
	})

	t.Run("GET presign", func(t *testing.T) {
		presigner := &fakeS3PresignerAPI{getErr: errors.New("presign get failed")}
		if _, err := newS3ObjectStore(&fakeS3ArchiveAPI{}, presigner, "private-bucket").PresignDownload(context.Background(), "archives/12345678-1234-4123-8123-123456789abc", 5*time.Minute); err == nil {
			t.Fatal("PresignDownload hid GET presign failure")
		}
	})

	t.Run("DELETE", func(t *testing.T) {
		client := &fakeS3ArchiveAPI{deleteErr: errors.New("delete failed")}
		if err := newS3ObjectStore(client, &fakeS3PresignerAPI{}, "private-bucket").DeleteObject(context.Background(), "archives/12345678-1234-4123-8123-123456789abc"); err == nil {
			t.Fatal("DeleteObject hid DeleteObject failure")
		}
	})
}

type fakeArchiveRepository struct {
	entries      map[string]ArchiveEntry
	listEntries  []ArchiveEntry
	createErr    error
	getErr       error
	listErr      error
	setStatusErr error
	deleteErr    error
}

func newFakeArchiveRepository() *fakeArchiveRepository {
	return &fakeArchiveRepository{entries: make(map[string]ArchiveEntry)}
}

func (f *fakeArchiveRepository) CreatePending(_ context.Context, entry ArchiveEntry) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.entries[entry.ID] = entry
	return nil
}

func (f *fakeArchiveRepository) Get(_ context.Context, id string) (ArchiveEntry, error) {
	if f.getErr != nil {
		return ArchiveEntry{}, f.getErr
	}
	entry, ok := f.entries[id]
	if !ok {
		return ArchiveEntry{}, ErrArchiveNotFound
	}
	return entry, nil
}

func (f *fakeArchiveRepository) ListAvailable(context.Context) ([]ArchiveEntry, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	entries := append([]ArchiveEntry(nil), f.listEntries...)
	if f.listEntries == nil {
		for _, entry := range f.entries {
			if entry.State == "available" {
				entries = append(entries, entry)
			}
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].UploadedAt > entries[j].UploadedAt })
	return entries, nil
}

func (f *fakeArchiveRepository) SetStatus(_ context.Context, id, state string) error {
	if f.setStatusErr != nil {
		return f.setStatusErr
	}
	entry, ok := f.entries[id]
	if !ok {
		return ErrArchiveNotFound
	}
	entry.State = state
	f.entries[id] = entry
	return nil
}

func (f *fakeArchiveRepository) Delete(_ context.Context, id string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.entries, id)
	return nil
}

type fakeArchiveObjectStore struct {
	createUploadCalls  int
	lastPendingKey     string
	lastContentType    string
	lastSize           int64
	lastExpiresIn      time.Duration
	createUploadErr    error
	verifyCalls        int
	verifyErr          error
	promoteCalls       int
	lastPromoteSource  string
	lastPromoteTarget  string
	promoteErr         error
	presignDownloadErr error
	deletedKeys        []string
	deleteErrByKey     map[string]error
}

func newFakeArchiveObjectStore() *fakeArchiveObjectStore {
	return &fakeArchiveObjectStore{deleteErrByKey: make(map[string]error)}
}

func (f *fakeArchiveObjectStore) CreateUpload(_ context.Context, pendingKey, contentType string, size int64, expiresIn time.Duration) (PresignedUpload, error) {
	f.createUploadCalls++
	f.lastPendingKey, f.lastContentType, f.lastSize, f.lastExpiresIn = pendingKey, contentType, size, expiresIn
	if f.createUploadErr != nil {
		return PresignedUpload{}, f.createUploadErr
	}
	return PresignedUpload{URL: "https://uploads.invalid", Fields: map[string]string{"key": pendingKey}, ExpiresAt: time.Now().Add(expiresIn)}, nil
}

func (f *fakeArchiveObjectStore) VerifyUpload(context.Context, string, string, string, int64) error {
	f.verifyCalls++
	return f.verifyErr
}

func (f *fakeArchiveObjectStore) PromoteUpload(_ context.Context, pendingKey, archiveKey string) error {
	f.promoteCalls++
	f.lastPromoteSource, f.lastPromoteTarget = pendingKey, archiveKey
	return f.promoteErr
}

func (f *fakeArchiveObjectStore) PresignDownload(_ context.Context, _ string, expiresIn time.Duration) (string, error) {
	if f.presignDownloadErr != nil {
		return "", f.presignDownloadErr
	}
	f.lastExpiresIn = expiresIn
	return "https://downloads.invalid", nil
}

func (f *fakeArchiveObjectStore) DeleteObject(_ context.Context, key string) error {
	f.deletedKeys = append(f.deletedKeys, key)
	for prefix, err := range f.deleteErrByKey {
		if strings.HasPrefix(key, prefix) {
			return err
		}
	}
	return nil
}

func validUploadMetadata() UploadMetadata {
	return UploadMetadata{
		OriginalName: "presentation.pdf",
		ContentType:  "application/pdf",
		Title:        "Presentation",
		Version:      "1.0",
		Date:         "2026-10-22",
		Summary:      "Quarterly presentation",
		Size:         8,
	}
}

func uploadIDFromFake(repository *fakeArchiveRepository) string {
	for id := range repository.entries {
		return id
	}
	return ""
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

type fakeDynamoDBAPI struct {
	putInput     *dynamodb.PutItemInput
	getOutput    *dynamodb.GetItemOutput
	queryInputs  []*dynamodb.QueryInput
	queryOutputs []*dynamodb.QueryOutput
	updateInput  *dynamodb.UpdateItemInput
	putErr       error
	getErr       error
	queryErr     error
	updateErr    error
	deleteErr    error
}

func (f *fakeDynamoDBAPI) PutItem(_ context.Context, input *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	f.putInput = input
	if f.putErr != nil {
		return nil, f.putErr
	}
	return &dynamodb.PutItemOutput{}, nil
}

func (f *fakeDynamoDBAPI) GetItem(_ context.Context, _ *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.getOutput != nil {
		return f.getOutput, nil
	}
	return &dynamodb.GetItemOutput{}, nil
}

func (f *fakeDynamoDBAPI) Query(_ context.Context, input *dynamodb.QueryInput, _ ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	f.queryInputs = append(f.queryInputs, input)
	if f.queryErr != nil {
		return nil, f.queryErr
	}
	if len(f.queryOutputs) == 0 {
		return &dynamodb.QueryOutput{}, nil
	}
	output := f.queryOutputs[0]
	f.queryOutputs = f.queryOutputs[1:]
	return output, nil
}

func (f *fakeDynamoDBAPI) UpdateItem(_ context.Context, input *dynamodb.UpdateItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	f.updateInput = input
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return &dynamodb.UpdateItemOutput{}, nil
}

func (f *fakeDynamoDBAPI) DeleteItem(_ context.Context, _ *dynamodb.DeleteItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	if f.deleteErr != nil {
		return nil, f.deleteErr
	}
	return &dynamodb.DeleteItemOutput{}, nil
}

func dynamoEntryAttributes(entry ArchiveEntry) map[string]types.AttributeValue {
	item, err := attributevalue.MarshalMap(entry)
	if err != nil {
		panic(err)
	}
	return item
}

type fakeS3ArchiveAPI struct {
	body          []byte
	contentType   string
	contentLength int64
	headErr       error
	getErr        error
	copyErr       error
	deleteErr     error
	headCalls     int
	getCalls      int
	copyCalls     int
	deleteCalls   int
	copySource    string
	copyTarget    string
	deletedKey    string
}

func (f *fakeS3ArchiveAPI) HeadObject(_ context.Context, _ *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	f.headCalls++
	if f.headErr != nil {
		return nil, f.headErr
	}
	size := f.contentLength
	if size == 0 && f.body != nil {
		size = int64(len(f.body))
	}
	return &s3.HeadObjectOutput{ContentLength: aws.Int64(size), ContentType: aws.String(f.contentType)}, nil
}

func (f *fakeS3ArchiveAPI) GetObject(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	f.getCalls++
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(f.body))}, nil
}

func (f *fakeS3ArchiveAPI) CopyObject(_ context.Context, input *s3.CopyObjectInput, _ ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
	f.copyCalls++
	f.copySource, f.copyTarget = aws.ToString(input.CopySource), aws.ToString(input.Key)
	if f.copyErr != nil {
		return nil, f.copyErr
	}
	return &s3.CopyObjectOutput{}, nil
}

func (f *fakeS3ArchiveAPI) DeleteObject(_ context.Context, input *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	f.deleteCalls++
	f.deletedKey = aws.ToString(input.Key)
	if f.deleteErr != nil {
		return nil, f.deleteErr
	}
	return &s3.DeleteObjectOutput{}, nil
}

type fakeS3PresignerAPI struct {
	postInput   *s3.PutObjectInput
	postOptions s3.PresignPostOptions
	postErr     error
	getInput    *s3.GetObjectInput
	getOptions  s3.PresignOptions
	getErr      error
}

func (f *fakeS3PresignerAPI) PresignPostObject(_ context.Context, input *s3.PutObjectInput, options ...func(*s3.PresignPostOptions)) (*s3.PresignedPostRequest, error) {
	f.postInput = input
	for _, option := range options {
		option(&f.postOptions)
	}
	if f.postErr != nil {
		return nil, f.postErr
	}
	return &s3.PresignedPostRequest{URL: "https://s3.invalid/upload", Values: map[string]string{"key": aws.ToString(input.Key), "Content-Type": aws.ToString(input.ContentType)}}, nil
}

func (f *fakeS3PresignerAPI) PresignGetObject(_ context.Context, input *s3.GetObjectInput, options ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	f.getInput = input
	for _, option := range options {
		option(&f.getOptions)
	}
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &v4.PresignedHTTPRequest{URL: "https://s3.invalid/download"}, nil
}

func hasPolicyCondition(conditions []any, expected []any) bool {
	for _, condition := range conditions {
		if reflect.DeepEqual(condition, expected) {
			return true
		}
	}
	return false
}
