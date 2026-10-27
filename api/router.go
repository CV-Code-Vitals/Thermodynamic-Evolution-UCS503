package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Dependencies struct {
	Auth       Authenticator
	Archives   ArchiveService
	LocalFiles LocalFileStore
	Logger     *slog.Logger
	Analysis   http.Handler
	Accounts   *githubAccountService
}

const archiveRouteTimeout = 25 * time.Second

func writeJSONError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}

func writeArchiveError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, operation string, err error) {
	status, code := http.StatusBadGateway, "storage_unavailable"
	switch {
	case errors.Is(err, ErrArchiveNotFound):
		status, code = http.StatusNotFound, "archive_not_found"
	case errors.Is(err, ErrInvalidUpload):
		status, code = http.StatusBadRequest, "invalid_upload"
	case errors.Is(err, ErrArchiveNotReady), errors.Is(err, ErrArchiveNotPending):
		status, code = http.StatusConflict, "archive_conflict"
	case errors.Is(err, context.DeadlineExceeded):
		status, code = http.StatusGatewayTimeout, "request_timeout"
	}
	if status >= http.StatusInternalServerError && logger != nil {
		logger.ErrorContext(r.Context(), "archive request failed",
			"operation", operation,
			"error_code", code,
			"status", status,
		)
	}
	writeJSONError(w, status, code)
}

func requestScopedTimeout(r *http.Request, timeout time.Duration) (*http.Request, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	return r.WithContext(ctx), cancel
}

func routeAuthorized(r *http.Request, auth Authenticator) bool {
	if auth != nil {
		const prefix = "Bearer "
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, prefix) || len(header) == len(prefix) {
			return false
		}
		return auth.Verify(strings.TrimSpace(header[len(prefix):]), time.Now())
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") {
		return false
	}
	return authorized(r)
}

func registerArchiveServiceRoutes(mux *http.ServeMux, deps Dependencies) {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	authorize := func(w http.ResponseWriter, r *http.Request) bool {
		if !routeAuthorized(r, deps.Auth) {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return false
		}
		return true
	}
	mux.HandleFunc("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		if !authorize(w, r) || !requireTrustedOrigin(w, r) {
			return
		}
		var metadata UploadMetadata
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(r.Body).Decode(&metadata); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_upload_metadata")
			return
		}
		if err := validateUploadMetadata(metadata.OriginalName, metadata.ContentType, metadata.Size); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_upload_metadata")
			return
		}
		if err := validateArchiveMetadata(metadata.Title, metadata.Version, metadata.Date, metadata.Summary); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_archive_metadata")
			return
		}
		callRequest, cancel := requestScopedTimeout(r, archiveRouteTimeout)
		defer cancel()
		upload, err := deps.Archives.BeginUpload(callRequest.Context(), metadata)
		if err != nil {
			writeArchiveError(w, r, logger, "begin_upload", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(upload)
	})
	mux.HandleFunc("/api/upload/complete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		if !authorize(w, r) || !requireTrustedOrigin(w, r) {
			return
		}
		var request struct {
			UploadID string `json:"uploadId"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.UploadID == "" {
			writeJSONError(w, http.StatusBadRequest, "invalid_upload_id")
			return
		}
		callRequest, cancel := requestScopedTimeout(r, archiveRouteTimeout)
		defer cancel()
		entry, err := deps.Archives.CompleteUpload(callRequest.Context(), request.UploadID)
		if err != nil {
			writeArchiveError(w, r, logger, "complete_upload", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(entry)
	})
	if deps.LocalFiles != nil {
		registerLocalFileRoutes(mux, deps)
	}
	mux.HandleFunc("/api/deliverables", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			writeJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		if !authorize(w, r) {
			return
		}
		callRequest, cancel := requestScopedTimeout(r, archiveRouteTimeout)
		defer cancel()
		entries, err := deps.Archives.List(callRequest.Context())
		if err != nil {
			writeArchiveError(w, r, logger, "list_archives", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(entries)
	})
	mux.HandleFunc("/api/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			writeJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		if !authorize(w, r) {
			return
		}
		id := r.URL.Query().Get("filename")
		if id == "" {
			writeJSONError(w, http.StatusBadRequest, "archive_id_required")
			return
		}
		callRequest, cancel := requestScopedTimeout(r, archiveRouteTimeout)
		defer cancel()
		download, err := deps.Archives.Download(callRequest.Context(), id)
		if err != nil {
			writeArchiveError(w, r, logger, "create_download_url", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(download)
	})
	mux.HandleFunc("/api/archive", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.Header().Set("Allow", http.MethodDelete)
			writeJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		if !authorize(w, r) || !requireTrustedOrigin(w, r) {
			return
		}
		id := r.URL.Query().Get("filename")
		if id == "" {
			writeJSONError(w, http.StatusBadRequest, "archive_id_required")
			return
		}
		callRequest, cancel := requestScopedTimeout(r, archiveRouteTimeout)
		defer cancel()
		if err := deps.Archives.Delete(callRequest.Context(), id); err != nil {
			writeArchiveError(w, r, logger, "delete_archive", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func newRouter(deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	registerAnalysisRoutes(mux, deps)
	registerAccountRoutes(mux, deps.Accounts)

	statusHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			writeJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := HealthResponse{Status: "SYS.INITIALIZE", Module: "Go Backend", Version: "1.0.0"}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("[ERROR] encode health response failed: %T", err)
		}
	}
	mux.HandleFunc("/status", statusHandler)
	mux.HandleFunc("/api/status", statusHandler)

	loginHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		if !requireTrustedOrigin(w, r) {
			return
		}

		var payload struct {
			Passkey string `json:"passkey"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			if err := r.ParseForm(); err == nil {
				payload.Passkey = r.FormValue("passkey")
			}
		}

		passkey := strings.TrimSpace(payload.Passkey)
		if deps.Auth != nil {
			token, expiresAt, err := deps.Auth.Issue(passkey)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			expiresInMs := expiresAt.Sub(time.Now().UTC()).Milliseconds()
			if expiresInMs < 0 {
				expiresInMs = 0
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":      "authenticated",
				"token":       token,
				"expiresAt":   expiresAt,
				"expiresInMs": expiresInMs,
			})
			return
		}
		if strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		expectedKey := strings.TrimSpace(adminPasskey)
		if passkey == "" || !constantTimeEquals(passkey, expectedKey) {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		token, err := randomSessionToken()
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "session_unavailable")
			return
		}
		sessionMu.Lock()
		if previous, err := r.Cookie(sessionCookie); err == nil {
			delete(sessionStore, previous.Value)
		}
		sessionStore[token] = time.Now().Add(sessionTTL)
		sessionMu.Unlock()

		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookie,
			Value:    token,
			HttpOnly: true,
			Path:     "/",
			SameSite: sessionSameSite(r),
			Secure:   requestIsHTTPS(r),
			MaxAge:   int(sessionTTL.Seconds()),
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "authenticated"})
	}
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/api/login", loginHandler)
	mux.HandleFunc("/api/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeJSONError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		if !requireTrustedOrigin(w, r) {
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1))
		if err != nil || len(body) != 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid_request")
			return
		}

		renewalAuth, ok := deps.Auth.(TokenRenewer)
		authorization := r.Header.Values("Authorization")
		if !ok || len(authorization) != 1 || !strings.HasPrefix(authorization[0], "Bearer ") {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		token := strings.TrimPrefix(authorization[0], "Bearer ")
		if token == "" || strings.TrimSpace(token) != token || strings.ContainsAny(token, " \t\r\n") {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		now := time.Now().UTC()
		newToken, expiresAt, err := renewalAuth.Renew(token, now)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":      "authenticated",
			"token":       newToken,
			"expiresAt":   expiresAt,
			"expiresInMs": expiresAt.Sub(now).Milliseconds(),
		})
	})

	if deps.Archives != nil {
		registerArchiveServiceRoutes(mux, deps)
		return securityHeaders(corsMiddleware(mux))
	}

	logoutHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		if !requireTrustedOrigin(w, r) {
			return
		}
		if cookie, err := r.Cookie(sessionCookie); err == nil && cookie.Value != "" {
			sessionMu.Lock()
			delete(sessionStore, cookie.Value)
			sessionMu.Unlock()
		}
		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookie,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
			SameSite: sessionSameSite(r),
			Secure:   requestIsHTTPS(r),
		})
		w.WriteHeader(http.StatusOK)
	}
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/api/logout", logoutHandler)

	uploadHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		if !routeAuthorized(r, deps.Auth) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if !requireTrustedOrigin(w, r) {
			return
		}

		const maxFileSize = 50 << 20
		r.Body = http.MaxBytesReader(w, r.Body, maxFileSize+(1<<20))
		if err := r.ParseMultipartForm(250 << 20); err != nil {
			http.Error(w, "Invalid multipart/form-data", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Missing file in request", http.StatusBadRequest)
			return
		}
		defer file.Close()

		title := strings.TrimSpace(r.FormValue("title"))
		version := strings.TrimSpace(r.FormValue("version"))
		date := strings.TrimSpace(r.FormValue("date"))
		summary := strings.TrimSpace(r.FormValue("summary"))
		contentType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
		if err != nil {
			http.Error(w, "Invalid file content type", http.StatusBadRequest)
			return
		}
		fileContents, err := io.ReadAll(io.LimitReader(file, maxArchiveSize+1))
		if err != nil {
			http.Error(w, "Unable to read upload", http.StatusBadRequest)
			return
		}
		if err := validateUpload(header.Filename, contentType, int64(len(fileContents)), fileContents); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := validateArchiveMetadata(title, version, date, summary); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !useS3 {
			if err := ensureUploadsDir(); err != nil {
				http.Error(w, "Server error: unable to create uploads directory", http.StatusInternalServerError)
				return
			}
		}

		storageKey, err := randomOpaqueArchiveID()
		if err != nil {
			http.Error(w, "Unable to generate archive ID", http.StatusInternalServerError)
			return
		}
		displayName := filepath.Base(header.Filename)
		displayName = strings.Map(func(r rune) rune {
			if r < 0x20 || r == 0x7f {
				return -1
			}
			return r
		}, displayName)
		outPath := filepath.Join(uploadsDir, storageKey)

		if useS3 {
			ctx := r.Context()
			if s3Client == nil {
				http.Error(w, "Server error: S3 is not configured", http.StatusInternalServerError)
				return
			}
			_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
				Bucket:      &s3Bucket,
				Key:         &storageKey,
				Body:        bytes.NewReader(fileContents),
				ContentType: &contentType,
			})
			if err != nil {
				log.Printf("[ERROR] unable to upload archive %q to S3: %v", storageKey, err)
				http.Error(w, "Server error: unable to upload to S3", http.StatusInternalServerError)
				return
			}
		} else {
			if err := os.WriteFile(outPath, fileContents, 0o600); err != nil {
				http.Error(w, "Server error: unable to write file", http.StatusInternalServerError)
				return
			}
		}

		entry := ArchiveEntry{
			ID:           storageKey,
			Filename:     storageKey,
			OriginalName: displayName,
			ContentType:  contentType,
			Size:         int64(len(fileContents)),
			Title:        title,
			Version:      version,
			Date:         date,
			Summary:      summary,
			UploadedAt:   time.Now().UTC().Format(time.RFC3339),
			State:        "available",
			URL:          archiveDownloadURL(storageKey),
		}
		archiveMutationMu.Lock()
		defer archiveMutationMu.Unlock()
		entries, err := loadArchives()
		if err != nil {
			if useS3 && s3Client != nil {
				if _, cleanupErr := s3Client.DeleteObject(r.Context(), &s3.DeleteObjectInput{Bucket: &s3Bucket, Key: &storageKey}); cleanupErr != nil {
					log.Printf("[ERROR] unable to clean up orphaned S3 object %q: %v", storageKey, cleanupErr)
				}
			} else if cleanupErr := os.Remove(outPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
				log.Printf("[ERROR] unable to clean up orphaned local file %q: %v", outPath, cleanupErr)
			}
			http.Error(w, "Server error: unable to load archives", http.StatusInternalServerError)
			return
		}

		entries = append([]ArchiveEntry{entry}, entries...)
		if err := saveArchives(entries); err != nil {
			if useS3 && s3Client != nil {
				if _, cleanupErr := s3Client.DeleteObject(r.Context(), &s3.DeleteObjectInput{Bucket: &s3Bucket, Key: &storageKey}); cleanupErr != nil {
					log.Printf("[ERROR] unable to clean up orphaned S3 object %q: %v", storageKey, cleanupErr)
				}
			} else if cleanupErr := os.Remove(outPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
				log.Printf("[ERROR] unable to clean up orphaned local file %q: %v", outPath, cleanupErr)
			}
			http.Error(w, "Server error: unable to save archives", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(entry)
	}
	mux.HandleFunc("/upload", uploadHandler)
	mux.HandleFunc("/api/upload", uploadHandler)
	routerDownloadHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		if !routeAuthorized(r, deps.Auth) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		serveDownload(w, r)
	}
	mux.HandleFunc("/download", routerDownloadHandler)
	mux.HandleFunc("/api/download", routerDownloadHandler)

	archivesHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		if !routeAuthorized(r, deps.Auth) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		entries, err := loadArchives()
		if err != nil {
			http.Error(w, "Server error: unable to read archives", http.StatusInternalServerError)
			return
		}
		for i := range entries {
			entries[i].URL = archiveDownloadURL(entries[i].Filename)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(entries)
	}
	mux.HandleFunc("/archives", archivesHandler)
	mux.HandleFunc("/api/archives", archivesHandler)
	mux.HandleFunc("/deliverables", archivesHandler)
	mux.HandleFunc("/api/deliverables", archivesHandler)

	archiveHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodDelete {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		if !routeAuthorized(r, deps.Auth) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if !requireTrustedOrigin(w, r) {
			return
		}

		requestedFilename := r.URL.Query().Get("filename")
		filename := filepath.Base(requestedFilename)
		if filename == "" || filename == "." || filename != requestedFilename {
			http.Error(w, "filename required", http.StatusBadRequest)
			return
		}
		archiveMutationMu.Lock()
		defer archiveMutationMu.Unlock()

		if useS3 {
			ctx := context.Background()
			_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &s3Bucket, Key: &filename})
			if err != nil {
				http.Error(w, "Unable to remove file from S3", http.StatusInternalServerError)
				return
			}
		} else {
			p := filepath.Join(uploadsDir, filename)
			if err := os.Remove(p); err != nil {
				http.Error(w, "Unable to remove file", http.StatusInternalServerError)
				return
			}
		}
		if useDB && dbPool != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			if _, err := dbPool.Exec(ctx, "DELETE FROM archives WHERE filename=$1", filename); err != nil {
				http.Error(w, "Unable to remove archive metadata", http.StatusInternalServerError)
				return
			}
		}

		entries, err := loadArchives()
		if err != nil {
			http.Error(w, "Server error: unable to read archives", http.StatusInternalServerError)
			return
		}
		filtered := make([]ArchiveEntry, 0, len(entries))
		for _, e := range entries {
			if e.Filename != filename {
				filtered = append(filtered, e)
			}
		}
		if err := saveArchives(filtered); err != nil {
			http.Error(w, "Server error: unable to save archives", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
	mux.HandleFunc("/archive", archiveHandler)
	mux.HandleFunc("/api/archive", archiveHandler)

	if !useS3 {
		fileHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !routeAuthorized(r, deps.Auth) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))).ServeHTTP(w, r)
		})
		mux.Handle("/uploads/", fileHandler)
		mux.Handle("/api/uploads/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !routeAuthorized(r, deps.Auth) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.StripPrefix("/api/uploads/", http.FileServer(http.Dir(uploadsDir))).ServeHTTP(w, r)
		}))
	}

	return securityHeaders(corsMiddleware(mux))
}
