package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLocalArchiveUploadFinalizeListDownloadDelete(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_REGION", "")

	handler, store, service := newTestLocalArchiveHandler(t)
	metadata := validUploadMetadata()
	contents := []byte("%PDF-1.7\nlocal test")
	metadata.Size = int64(len(contents))
	startedAt := time.Now()
	upload := beginLocalUpload(t, handler, metadata)
	if upload.UploadID == "" {
		t.Fatal("local upload response did not include an opaque upload ID")
	}
	if !strings.HasPrefix(upload.URL, "/api/local-upload?") {
		t.Fatalf("local upload URL %q does not target the local API", upload.URL)
	}
	lifetime := upload.ExpiresAt.Sub(startedAt)
	if lifetime < archiveURLLifetime-time.Second || lifetime > archiveURLLifetime+time.Second {
		t.Fatalf("local upload authorization lifetime = %s, want 5 minutes", lifetime)
	}

	uploadResponse := localUploadRequest(handler, upload.URL, metadata.ContentType, contents, "valid-bearer")
	if uploadResponse.Code != http.StatusNoContent {
		t.Fatalf("local upload returned %d: %s", uploadResponse.Code, uploadResponse.Body.String())
	}
	if err := store.VerifyUpload(context.Background(), "pending/"+upload.UploadID, metadata.OriginalName, metadata.ContentType, int64(len(contents))); err != nil {
		t.Fatalf("shared finalize verification rejected staged file: %v", err)
	}

	completeResponse := authorizedRequest(handler, http.MethodPost, "/api/upload/complete", `{"uploadId":"`+upload.UploadID+`"}`)
	if completeResponse.Code != http.StatusCreated {
		_, finalizeErr := service.CompleteUpload(context.Background(), upload.UploadID)
		t.Fatalf("finalize returned %d: %s; service error: %v", completeResponse.Code, completeResponse.Body.String(), finalizeErr)
	}
	var entry ArchiveEntry
	if err := json.Unmarshal(completeResponse.Body.Bytes(), &entry); err != nil {
		t.Fatalf("decode finalized archive: %v", err)
	}
	if entry.State != "available" || entry.ID != upload.UploadID {
		t.Fatalf("unexpected finalized entry: %#v", entry)
	}
	reused := localUploadRequest(handler, upload.URL, metadata.ContentType, contents, "valid-bearer")
	if reused.Code != http.StatusGone {
		t.Fatalf("finalized upload ID reuse returned %d, want %d", reused.Code, http.StatusGone)
	}

	listResponse := authorizedRequest(handler, http.MethodGet, "/api/deliverables", "")
	if listResponse.Code != http.StatusOK {
		t.Fatalf("list returned %d: %s", listResponse.Code, listResponse.Body.String())
	}
	var entries []ArchiveEntry
	if err := json.Unmarshal(listResponse.Body.Bytes(), &entries); err != nil {
		t.Fatalf("decode archive list: %v", err)
	}
	if len(entries) != 1 || entries[0].ID != entry.ID {
		t.Fatalf("list did not return the finalized archive: %#v", entries)
	}

	downloadResponse := authorizedRequest(handler, http.MethodGet, "/api/download?filename="+url.QueryEscape(entry.ID), "")
	if downloadResponse.Code != http.StatusOK {
		t.Fatalf("download URL request returned %d: %s", downloadResponse.Code, downloadResponse.Body.String())
	}
	var download PresignedDownload
	if err := json.Unmarshal(downloadResponse.Body.Bytes(), &download); err != nil {
		t.Fatalf("decode download response: %v", err)
	}
	downloadResponse = authorizedRequest(handler, http.MethodGet, download.URL, "")
	if downloadResponse.Code != http.StatusOK || !bytes.Equal(downloadResponse.Body.Bytes(), contents) {
		t.Fatalf("local file download returned %d with body %q", downloadResponse.Code, downloadResponse.Body.Bytes())
	}

	deleteResponse := authorizedRequest(handler, http.MethodDelete, "/api/archive?filename="+url.QueryEscape(entry.ID), "")
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("delete returned %d: %s", deleteResponse.Code, deleteResponse.Body.String())
	}
	listResponse = authorizedRequest(handler, http.MethodGet, "/api/deliverables", "")
	if listResponse.Code != http.StatusOK {
		t.Fatalf("list after delete returned %d: %s", listResponse.Code, listResponse.Body.String())
	}
	if err := json.Unmarshal(listResponse.Body.Bytes(), &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("deleted archive remains listed: %#v", entries)
	}
}

func TestLocalArchiveListExcludesLegacyRecordsWithoutMigratingThem(t *testing.T) {
	root := t.TempDir()
	archivesPath := filepath.Join(root, "archives.json")
	legacy := []byte(`[{"filename":"old-report.pdf","originalName":"old-report.pdf","state":"available","title":"Legacy"}]`)
	if err := os.WriteFile(archivesPath, legacy, 0o600); err != nil {
		t.Fatal(err)
	}
	repository := newLocalArchiveRepository(archivesPath)
	service := newArchiveService(repository, newLocalObjectStore(filepath.Join(root, "uploads")))

	entries, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("list legacy local archives: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("legacy archive without a new-store object key was listed: %#v", entries)
	}
	persisted, err := os.ReadFile(archivesPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(persisted, legacy) {
		t.Fatal("listing legacy archives modified or migrated their metadata")
	}
}

func TestLocalUploadRequiresBearerAuthorization(t *testing.T) {
	handler, _, _ := newTestLocalArchiveHandler(t)
	upload := beginLocalUpload(t, handler, validUploadMetadata())

	request := httptest.NewRequest(http.MethodPost, upload.URL, strings.NewReader("%PDF-1.7"))
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Content-Type", "application/pdf")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized local upload returned %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestLocalUploadRejectsMissingExpiredAndReusedIDs(t *testing.T) {
	handler, store, _ := newTestLocalArchiveHandler(t)
	missing := localUploadRequest(handler, "/api/local-upload?uploadId=missing", "application/pdf", []byte("%PDF-1.7"), "valid-bearer")
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing upload ID returned %d, want %d", missing.Code, http.StatusNotFound)
	}

	upload := beginLocalUpload(t, handler, validUploadMetadata())
	store.now = func() time.Time { return upload.ExpiresAt.Add(time.Nanosecond) }
	expired := localUploadRequest(handler, upload.URL, "application/pdf", []byte("%PDF-1.7"), "valid-bearer")
	if expired.Code != http.StatusGone {
		t.Fatalf("expired upload ID returned %d, want %d: %s", expired.Code, http.StatusGone, expired.Body.String())
	}

	store.now = time.Now
	upload = beginLocalUpload(t, handler, validUploadMetadata())
	first := localUploadRequest(handler, upload.URL, "application/pdf", []byte("%PDF-1.7"), "valid-bearer")
	if first.Code != http.StatusNoContent {
		t.Fatalf("first use returned %d: %s", first.Code, first.Body.String())
	}
	reused := localUploadRequest(handler, upload.URL, "application/pdf", []byte("%PDF-1.7"), "valid-bearer")
	if reused.Code != http.StatusGone {
		t.Fatalf("reused upload ID returned %d, want %d", reused.Code, http.StatusGone)
	}
}

func TestLocalUploadRejectsOversizedBodyAndInvalidSignature(t *testing.T) {
	handler, _, _ := newTestLocalArchiveHandler(t)
	upload := beginLocalUpload(t, handler, validUploadMetadata())
	oversized := localUploadRequest(handler, upload.URL, "application/pdf", bytes.Repeat([]byte("x"), int(maxArchiveSize+1)), "valid-bearer")
	if oversized.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized local upload returned %d, want %d", oversized.Code, http.StatusRequestEntityTooLarge)
	}

	metadata := validUploadMetadata()
	metadata.Size = int64(len("not a PDF"))
	upload = beginLocalUpload(t, handler, metadata)
	invalid := localUploadRequest(handler, upload.URL, "application/pdf", []byte("not a PDF"), "valid-bearer")
	if invalid.Code != http.StatusNoContent {
		t.Fatalf("staging invalid signature returned %d, want %d: %s", invalid.Code, http.StatusNoContent, invalid.Body.String())
	}
	complete := authorizedRequest(handler, http.MethodPost, "/api/upload/complete", `{"uploadId":"`+upload.UploadID+`"}`)
	if complete.Code == http.StatusCreated {
		t.Fatal("shared finalize validation exposed an upload with an invalid signature")
	}
}

func TestLocalUploadSurfacesFilesystemWriteFailure(t *testing.T) {
	handler, store, _ := newTestLocalArchiveHandler(t)
	upload := beginLocalUpload(t, handler, validUploadMetadata())
	blockingPath := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blockingPath, []byte("block"), 0o600); err != nil {
		t.Fatal(err)
	}
	store.root = blockingPath

	response := localUploadRequest(handler, upload.URL, "application/pdf", []byte("%PDF-1.7"), "valid-bearer")
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("filesystem failure returned %d, want %d: %s", response.Code, http.StatusInternalServerError, response.Body.String())
	}
	if strings.Contains(strings.ToLower(response.Body.String()), "uploaded") {
		t.Fatalf("filesystem failure was reported as successful: %s", response.Body.String())
	}
}

func newTestLocalArchiveHandler(t *testing.T) (http.Handler, *LocalObjectStore, *archiveService) {
	t.Helper()
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	dir := t.TempDir()
	repository := newLocalArchiveRepository(filepath.Join(dir, "archives.json"))
	store := newLocalObjectStore(filepath.Join(dir, "uploads"))
	service := newArchiveService(repository, store)
	auth := &testAuthenticator{validToken: "valid-bearer"}
	return newRouter(Dependencies{Auth: auth, Archives: service, LocalFiles: store}), store, service
}

func beginLocalUpload(t *testing.T, handler http.Handler, metadata UploadMetadata) PresignedUpload {
	t.Helper()
	body, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/upload", bytes.NewReader(body))
	request.Header.Set("Authorization", "Bearer valid-bearer")
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("begin upload returned %d: %s", response.Code, response.Body.String())
	}
	var upload PresignedUpload
	if err := json.Unmarshal(response.Body.Bytes(), &upload); err != nil {
		t.Fatalf("decode upload authorization: %v", err)
	}
	return upload
}

func localUploadRequest(handler http.Handler, path, contentType string, contents []byte, token string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(contents))
	request.Header.Set("Content-Type", contentType)
	request.Header.Set("Origin", "http://localhost:5173")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func authorizedRequest(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer valid-bearer")
	request.Header.Set("Origin", "http://localhost:5173")
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func TestLocalRepositoryReturnsFilesystemErrors(t *testing.T) {
	blockingPath := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blockingPath, []byte("block"), 0o600); err != nil {
		t.Fatal(err)
	}
	repository := newLocalArchiveRepository(filepath.Join(blockingPath, "archives.json"))
	err := repository.CreatePending(context.Background(), ArchiveEntry{
		ID: "4ac01d16-e566-42ab-b9c4-d50b8bd216c1", State: "pending",
		PendingKey: "pending/4ac01d16-e566-42ab-b9c4-d50b8bd216c1",
		ObjectKey:  "archives/4ac01d16-e566-42ab-b9c4-d50b8bd216c1",
		ExpiresAt:  time.Now().Add(time.Hour).Unix(),
	})
	if err == nil || errors.Is(err, ErrArchiveNotFound) {
		t.Fatalf("repository did not surface filesystem error: %v", err)
	}
}

func TestLocalFinalizeRetriesAfterMetadataUpdateFailure(t *testing.T) {
	root := t.TempDir()
	repository := &failOnceLocalStatusRepository{
		localArchiveRepository: newLocalArchiveRepository(filepath.Join(root, "archives.json")),
		failAvailable:          true,
	}
	objects := newLocalObjectStore(filepath.Join(root, "uploads"))
	service := newArchiveService(repository, objects)
	contents := []byte("%PDF-1.7\nlocal retry")
	metadata := validUploadMetadata()
	metadata.Size = int64(len(contents))
	upload, err := service.BeginUpload(context.Background(), metadata)
	if err != nil {
		t.Fatal(err)
	}
	if err := objects.StageUpload(context.Background(), upload.UploadID, metadata.ContentType, bytes.NewReader(contents)); err != nil {
		t.Fatal(err)
	}

	if _, err := service.CompleteUpload(context.Background(), upload.UploadID); err == nil {
		t.Fatal("first completion did not surface the simulated metadata update failure")
	}
	entry, err := repository.Get(context.Background(), upload.UploadID)
	if err != nil {
		t.Fatal(err)
	}
	if entry.State != "pending" {
		t.Fatalf("failed metadata update changed persisted state to %q", entry.State)
	}
	if _, err := os.Stat(filepath.Join(root, "uploads", filepath.FromSlash(entry.ObjectKey))); err != nil {
		t.Fatalf("failed metadata update discarded the promoted copy: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "uploads", filepath.FromSlash(entry.PendingKey))); err != nil {
		t.Fatalf("failed metadata update discarded the staging object: %v", err)
	}

	finalized, err := service.CompleteUpload(context.Background(), upload.UploadID)
	if err != nil {
		t.Fatalf("retry did not complete the upload: %v", err)
	}
	if finalized.State != "available" {
		t.Fatalf("retry returned state %q, want available", finalized.State)
	}
}

func TestLocalUploadRequestDoesNotSilentlyAcceptOversizeContentLength(t *testing.T) {
	handler, _, _ := newTestLocalArchiveHandler(t)
	upload := beginLocalUpload(t, handler, validUploadMetadata())
	request := httptest.NewRequest(http.MethodPost, upload.URL, strings.NewReader("%PDF-1.7"))
	request.ContentLength = maxArchiveSize + 1
	request.Header.Set("Content-Type", "application/pdf")
	request.Header.Set("Authorization", "Bearer valid-bearer")
	request.Header.Set("Origin", "http://localhost:5173")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized declared body returned %d, want %d", response.Code, http.StatusRequestEntityTooLarge)
	}

}

func TestLocalUploadMetadataIsPersistedAtConfiguredPath(t *testing.T) {
	dir := t.TempDir()
	archivesPath := filepath.Join(dir, "configured-archives.json")
	repository := newLocalArchiveRepository(archivesPath)
	entry := ArchiveEntry{
		ID: "4ac01d16-e566-42ab-b9c4-d50b8bd216c1", OriginalName: "report.pdf",
		ContentType: "application/pdf", Size: 8, Title: "Report", Version: "1.0",
		Date: "2026-10-22", Summary: "Summary", UploadedAt: time.Now().UTC().Format(time.RFC3339),
		State: "pending", PendingKey: "pending/4ac01d16-e566-42ab-b9c4-d50b8bd216c1",
		ObjectKey: "archives/4ac01d16-e566-42ab-b9c4-d50b8bd216c1", ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}

	if err := repository.CreatePending(context.Background(), entry); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(archivesPath); err != nil {
		t.Fatalf("local metadata was not written to configured archive path: %v", err)
	}
	loaded, err := repository.Get(context.Background(), entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ID != entry.ID || loaded.State != "pending" {
		t.Fatalf("unexpected persisted local metadata: %#v", loaded)
	}
}

type failOnceLocalStatusRepository struct {
	*localArchiveRepository
	failAvailable bool
}

func (r *failOnceLocalStatusRepository) SetStatus(ctx context.Context, id, state string) error {
	if state == "available" && r.failAvailable {
		r.failAvailable = false
		return errors.New("simulated metadata update failure")
	}
	return r.localArchiveRepository.SetStatus(ctx, id, state)
}
