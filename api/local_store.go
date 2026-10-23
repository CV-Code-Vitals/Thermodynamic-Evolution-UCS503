package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	ErrLocalUploadNotFound = errors.New("local upload authorization not found")
	ErrLocalUploadExpired  = errors.New("local upload authorization expired")
	ErrLocalUploadConsumed = errors.New("local upload authorization already used")
	ErrLocalUploadTooLarge = errors.New("local upload exceeds the maximum size")
)

type LocalFileStore interface {
	StageUpload(context.Context, string, string, io.Reader) error
	OpenArchive(context.Context, string) (*os.File, error)
}

type localUploadAuthorization struct {
	contentType string
	expiresAt   time.Time
	consumed    bool
}

type LocalObjectStore struct {
	root           string
	mu             sync.Mutex
	authorizations map[string]localUploadAuthorization
	now            func() time.Time
}

func newLocalObjectStore(root string) *LocalObjectStore {
	return &LocalObjectStore{
		root:           root,
		authorizations: make(map[string]localUploadAuthorization),
		now:            time.Now,
	}
}

func (s *LocalObjectStore) CreateUpload(_ context.Context, pendingKey, contentType string, size int64, expiresIn time.Duration) (PresignedUpload, error) {
	if !validGeneratedObjectKey(pendingKey, "pending/") {
		return PresignedUpload{}, errors.New("upload key must be a generated pending/ key")
	}
	if err := validateUploadMetadata("upload"+extensionForContentType(contentType), contentType, size); err != nil {
		return PresignedUpload{}, err
	}
	if expiresIn <= 0 || expiresIn > archiveURLLifetime {
		return PresignedUpload{}, errors.New("upload authorization expiry must be between 0 and 5 minutes")
	}
	id := strings.TrimPrefix(pendingKey, "pending/")
	expiresAt := s.now().UTC().Add(expiresIn)
	s.mu.Lock()
	for issuedID, authorization := range s.authorizations {
		if !authorization.expiresAt.After(s.now()) {
			delete(s.authorizations, issuedID)
		}
	}
	if _, exists := s.authorizations[id]; exists {
		s.mu.Unlock()
		return PresignedUpload{}, errors.New("local upload authorization already exists")
	}
	s.authorizations[id] = localUploadAuthorization{contentType: contentType, expiresAt: expiresAt}
	s.mu.Unlock()

	return PresignedUpload{
		URL:       "/api/local-upload?uploadId=" + url.QueryEscape(id),
		ExpiresAt: expiresAt,
	}, nil
}

func (s *LocalObjectStore) StageUpload(ctx context.Context, id, contentType string, body io.Reader) error {
	if !validGeneratedObjectKey("pending/"+id, "pending/") {
		return ErrLocalUploadNotFound
	}
	s.mu.Lock()
	authorization, exists := s.authorizations[id]
	switch {
	case !exists:
		s.mu.Unlock()
		return ErrLocalUploadNotFound
	case !authorization.expiresAt.After(s.now()):
		delete(s.authorizations, id)
		s.mu.Unlock()
		return ErrLocalUploadExpired
	case authorization.consumed:
		s.mu.Unlock()
		return ErrLocalUploadConsumed
	case authorization.contentType != contentType:
		s.mu.Unlock()
		return ErrInvalidUpload
	}
	authorization.consumed = true
	s.authorizations[id] = authorization
	s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("request canceled before local upload: %w", err)
	}
	pendingDir := filepath.Join(s.root, "pending")
	if err := os.MkdirAll(pendingDir, 0o755); err != nil {
		return fmt.Errorf("create local staging directory: %w", err)
	}
	staged, err := os.CreateTemp(pendingDir, id+"-*.stage")
	if err != nil {
		return fmt.Errorf("create unique local staging file: %w", err)
	}
	stagedPath := staged.Name()
	removeStaged := true
	defer func() {
		if removeStaged {
			_ = os.Remove(stagedPath)
		}
	}()
	written, err := io.Copy(staged, io.LimitReader(body, maxArchiveSize+1))
	if err != nil {
		_ = staged.Close()
		return fmt.Errorf("write local staging file: %w", err)
	}
	if written > maxArchiveSize {
		_ = staged.Close()
		return ErrLocalUploadTooLarge
	}
	if err := staged.Sync(); err != nil {
		_ = staged.Close()
		return fmt.Errorf("sync local staging file: %w", err)
	}
	if err := staged.Close(); err != nil {
		return fmt.Errorf("close local staging file: %w", err)
	}
	target := filepath.Join(pendingDir, id)
	if _, err := os.Lstat(target); err == nil {
		return ErrLocalUploadConsumed
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check local staging destination: %w", err)
	}
	if err := os.Rename(stagedPath, target); err != nil {
		return fmt.Errorf("publish local staging file: %w", err)
	}
	removeStaged = false
	return nil
}

func (s *LocalObjectStore) VerifyUpload(_ context.Context, pendingKey, originalName, contentType string, expectedSize int64) error {
	if !validGeneratedObjectKey(pendingKey, "pending/") {
		return errors.New("upload key must be a generated pending/ key")
	}
	if err := validateUploadMetadata(originalName, contentType, expectedSize); err != nil {
		return err
	}
	id := strings.TrimPrefix(pendingKey, "pending/")
	s.mu.Lock()
	authorization, exists := s.authorizations[id]
	s.mu.Unlock()
	if !exists || !authorization.consumed || authorization.contentType != contentType {
		return ErrInvalidUpload
	}
	file, err := os.Open(filepath.Join(s.root, "pending", id))
	if err != nil {
		return fmt.Errorf("open staged local upload: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxArchiveSize+1))
	if err != nil {
		return fmt.Errorf("read staged local upload: %w", err)
	}
	if int64(len(data)) != expectedSize {
		return ErrInvalidUpload
	}
	if err := validateUpload(originalName, contentType, expectedSize, data); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidUpload, err)
	}
	return nil
}

func (s *LocalObjectStore) PromoteUpload(_ context.Context, pendingKey, archiveKey string) error {
	if !validGeneratedObjectKey(pendingKey, "pending/") || !validGeneratedObjectKey(archiveKey, "archives/") {
		return errors.New("promotion requires generated pending/ and archives/ keys")
	}
	archivesDir := filepath.Join(s.root, "archives")
	if err := os.MkdirAll(archivesDir, 0o755); err != nil {
		return fmt.Errorf("create local archive directory: %w", err)
	}
	source := filepath.Join(s.root, filepath.FromSlash(pendingKey))
	target := filepath.Join(s.root, filepath.FromSlash(archiveKey))
	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open staged local archive: %w", err)
	}
	defer sourceFile.Close()
	targetFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if os.IsExist(err) {
			if same, compareErr := localFilesEqual(source, target); compareErr != nil {
				return compareErr
			} else if same {
				return nil
			}
			return errors.New("local archive destination already contains different content")
		}
		return fmt.Errorf("create local archive destination: %w", err)
	}
	removeTarget := true
	defer func() {
		if removeTarget {
			_ = os.Remove(target)
		}
	}()
	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		_ = targetFile.Close()
		return fmt.Errorf("copy staged local archive: %w", err)
	}
	if err := targetFile.Sync(); err != nil {
		_ = targetFile.Close()
		return fmt.Errorf("sync promoted local archive: %w", err)
	}
	if err := targetFile.Close(); err != nil {
		return fmt.Errorf("close promoted local archive: %w", err)
	}
	removeTarget = false
	return nil
}

func localFilesEqual(firstPath, secondPath string) (bool, error) {
	first, err := os.ReadFile(firstPath)
	if err != nil {
		return false, fmt.Errorf("read staged local archive for retry: %w", err)
	}
	second, err := os.ReadFile(secondPath)
	if err != nil {
		return false, fmt.Errorf("read existing local archive for retry: %w", err)
	}
	return bytes.Equal(first, second), nil
}

func (s *LocalObjectStore) PresignDownload(_ context.Context, archiveKey string, expiresIn time.Duration) (string, error) {
	if !validGeneratedObjectKey(archiveKey, "archives/") {
		return "", errors.New("download key must be a generated archives/ key")
	}
	if expiresIn <= 0 || expiresIn > archiveURLLifetime {
		return "", errors.New("download URL expiry must be between 0 and 5 minutes")
	}
	return "/api/local-download?key=" + url.QueryEscape(archiveKey), nil
}

func (s *LocalObjectStore) DeleteObject(_ context.Context, key string) error {
	if !validGeneratedObjectKey(key, "pending/") && !validGeneratedObjectKey(key, "archives/") {
		return errors.New("delete key must be under pending/ or archives/")
	}
	path := filepath.Join(s.root, filepath.FromSlash(key))
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete local archive object: %w", err)
	}
	return nil
}

func (s *LocalObjectStore) OpenArchive(_ context.Context, key string) (*os.File, error) {
	if !validGeneratedObjectKey(key, "archives/") {
		return nil, errors.New("download key must be a generated archives/ key")
	}
	file, err := os.Open(filepath.Join(s.root, filepath.FromSlash(key)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrArchiveNotFound
		}
		return nil, fmt.Errorf("open local archive object: %w", err)
	}
	return file, nil
}

type localArchiveRepository struct {
	path string
	mu   sync.Mutex
}

type localArchiveRecord struct {
	ArchiveEntry
	PendingKey string `json:"pendingKey,omitempty"`
	ObjectKey  string `json:"objectKey"`
	ExpiresAt  int64  `json:"expiresAt,omitempty"`
}

func newLocalArchiveRepository(path string) *localArchiveRepository {
	return &localArchiveRepository{path: path}
}

func (r *localArchiveRepository) CreatePending(_ context.Context, entry ArchiveEntry) error {
	if entry.ID == "" || entry.State != "pending" || entry.ExpiresAt <= 0 ||
		!validGeneratedObjectKey(entry.PendingKey, "pending/") ||
		!validGeneratedObjectKey(entry.ObjectKey, "archives/") {
		return errors.New("pending archive record has invalid state or object keys")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	entries, err := r.read()
	if err != nil {
		return err
	}
	if _, exists := entries[entry.ID]; exists {
		return errors.New("archive record already exists")
	}
	entries[entry.ID] = entry
	return r.write(entries)
}

func (r *localArchiveRepository) Get(_ context.Context, id string) (ArchiveEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entries, err := r.read()
	if err != nil {
		return ArchiveEntry{}, err
	}
	entry, exists := entries[id]
	if !exists {
		return ArchiveEntry{}, ErrArchiveNotFound
	}
	return entry, nil
}

func (r *localArchiveRepository) ListAvailable(_ context.Context) ([]ArchiveEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entries, err := r.read()
	if err != nil {
		return nil, err
	}
	available := make([]ArchiveEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.State == "available" {
			available = append(available, entry)
		}
	}
	return available, nil
}

func (r *localArchiveRepository) SetStatus(_ context.Context, id, state string) error {
	if state != "pending" && state != "available" && state != "deleting" {
		return fmt.Errorf("unsupported archive state %q", state)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	entries, err := r.read()
	if err != nil {
		return err
	}
	entry, exists := entries[id]
	if !exists {
		return ErrArchiveNotFound
	}
	entry.State = state
	if state == "available" {
		entry.ExpiresAt = 0
	}
	entries[id] = entry
	return r.write(entries)
}

func (r *localArchiveRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	entries, err := r.read()
	if err != nil {
		return err
	}
	if _, exists := entries[id]; !exists {
		return ErrArchiveNotFound
	}
	delete(entries, id)
	return r.write(entries)
}

func (r *localArchiveRepository) read() (map[string]ArchiveEntry, error) {
	file, err := os.Open(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]ArchiveEntry{}, nil
		}
		return nil, fmt.Errorf("open local archive metadata: %w", err)
	}
	defer file.Close()
	var raw json.RawMessage
	if err := json.NewDecoder(file).Decode(&raw); err != nil {
		if errors.Is(err, io.EOF) {
			return map[string]ArchiveEntry{}, nil
		}
		return nil, fmt.Errorf("decode local archive metadata: %w", err)
	}
	entries := make(map[string]ArchiveEntry)
	switch {
	case len(raw) == 0:
		return entries, nil
	case raw[0] == '[':
		var legacy []ArchiveEntry
		if err := json.Unmarshal(raw, &legacy); err != nil {
			return nil, fmt.Errorf("decode local archive metadata: %w", err)
		}
		for _, entry := range legacy {
			if entry.ID == "" {
				entry.ID = entry.Filename
			}
			if entry.ID != "" {
				entries[entry.ID] = entry
			}
		}
	default:
		var records map[string]localArchiveRecord
		if err := json.Unmarshal(raw, &records); err != nil {
			return nil, fmt.Errorf("decode local archive metadata: %w", err)
		}
		for id, record := range records {
			entry := record.ArchiveEntry
			entry.ID = id
			entry.PendingKey = record.PendingKey
			entry.ObjectKey = record.ObjectKey
			entry.ExpiresAt = record.ExpiresAt
			entries[id] = entry
		}
	}
	return entries, nil
}

func (r *localArchiveRepository) write(entries map[string]ArchiveEntry) error {
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create local archive metadata directory: %w", err)
	}
	temporary, err := os.CreateTemp(dir, ".archives-*.json")
	if err != nil {
		return fmt.Errorf("create temporary local archive metadata: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	records := make(map[string]localArchiveRecord, len(entries))
	for id, entry := range entries {
		records[id] = localArchiveRecord{
			ArchiveEntry: entry,
			PendingKey:   entry.PendingKey,
			ObjectKey:    entry.ObjectKey,
			ExpiresAt:    entry.ExpiresAt,
		}
	}
	if err := json.NewEncoder(temporary).Encode(records); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("encode local archive metadata: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync local archive metadata: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary local archive metadata: %w", err)
	}
	if err := replaceLocalFile(temporaryPath, r.path); err != nil {
		return fmt.Errorf("replace local archive metadata: %w", err)
	}
	return nil
}

func replaceLocalFile(source, destination string) error {
	if err := os.Rename(source, destination); err == nil {
		return nil
	} else if _, statErr := os.Stat(destination); os.IsNotExist(statErr) {
		return err
	}

	dir := filepath.Dir(destination)
	backup, err := os.CreateTemp(dir, ".archives-backup-*.json")
	if err != nil {
		return fmt.Errorf("create local metadata backup: %w", err)
	}
	backupPath := backup.Name()
	if err := backup.Close(); err != nil {
		_ = os.Remove(backupPath)
		return fmt.Errorf("close local metadata backup: %w", err)
	}
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("prepare local metadata backup: %w", err)
	}
	if err := os.Rename(destination, backupPath); err != nil {
		return fmt.Errorf("preserve existing local metadata: %w", err)
	}
	if err := os.Rename(source, destination); err != nil {
		restoreErr := os.Rename(backupPath, destination)
		return errors.Join(fmt.Errorf("install local metadata: %w", err), wrapOptionalError("restore local metadata", restoreErr))
	}
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("remove replaced local metadata backup: %w", err)
	}
	return nil
}

func registerLocalFileRoutes(mux *http.ServeMux, deps Dependencies) {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	mux.HandleFunc("/api/local-upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		if !routeAuthorized(r, deps.Auth) {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if !requireTrustedOrigin(w, r) {
			return
		}
		id := r.URL.Query().Get("uploadId")
		if !validGeneratedObjectKey("pending/"+id, "pending/") {
			writeJSONError(w, http.StatusNotFound, "upload_authorization_not_found")
			return
		}
		if r.ContentLength > maxArchiveSize {
			writeJSONError(w, http.StatusRequestEntityTooLarge, "upload_too_large")
			return
		}
		contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_upload_content_type")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxArchiveSize)
		callRequest, cancel := requestScopedTimeout(r, archiveRouteTimeout)
		defer cancel()
		if err := deps.LocalFiles.StageUpload(callRequest.Context(), id, contentType, r.Body); err != nil {
			var maxBytesErr *http.MaxBytesError
			switch {
			case errors.Is(err, ErrLocalUploadNotFound):
				writeJSONError(w, http.StatusNotFound, "upload_authorization_not_found")
			case errors.Is(err, ErrLocalUploadExpired), errors.Is(err, ErrLocalUploadConsumed):
				writeJSONError(w, http.StatusGone, "upload_authorization_expired")
			case errors.Is(err, ErrLocalUploadTooLarge), errors.As(err, &maxBytesErr):
				writeJSONError(w, http.StatusRequestEntityTooLarge, "upload_too_large")
			case errors.Is(err, ErrInvalidUpload):
				writeJSONError(w, http.StatusBadRequest, "upload_content_type_mismatch")
			default:
				logger.ErrorContext(r.Context(), "local archive upload failed", "error_code", "local_storage_unavailable")
				writeJSONError(w, http.StatusInternalServerError, "local_storage_unavailable")
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/local-download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			writeJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		if !routeAuthorized(r, deps.Auth) {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		callRequest, cancel := requestScopedTimeout(r, archiveRouteTimeout)
		defer cancel()
		file, err := deps.LocalFiles.OpenArchive(callRequest.Context(), r.URL.Query().Get("key"))
		if err != nil {
			if errors.Is(err, ErrArchiveNotFound) {
				writeJSONError(w, http.StatusNotFound, "archive_not_found")
				return
			}
			logger.ErrorContext(r.Context(), "local archive download failed", "error_code", "local_storage_unavailable")
			writeJSONError(w, http.StatusInternalServerError, "local_storage_unavailable")
			return
		}
		defer file.Close()
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="archive"`)
		if _, err := io.Copy(w, file); err != nil {
			logger.ErrorContext(r.Context(), "local archive download response failed", "error_code", "response_write_failed")
		}
	})
}

var (
	_ ArchiveRepository = (*localArchiveRepository)(nil)
	_ ObjectStore       = (*LocalObjectStore)(nil)
	_ LocalFileStore    = (*LocalObjectStore)(nil)
)
