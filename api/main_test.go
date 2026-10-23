package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type failingRandomReader struct{}

func (failingRandomReader) Read([]byte) (int, error) {
	return 0, errors.New("simulated entropy source failure")
}

type testS3Client struct {
	getObjectOutput *s3.GetObjectOutput
	getObjectError  error
	getObjectCalls  int
}

type testAuthenticator struct {
	issuedPasskey string
	issueToken    string
	issueExpiry   time.Time
	issueError    error
	verifyCalls   []string
	validToken    string
}

func (a *testAuthenticator) Issue(passkey string) (string, time.Time, error) {
	a.issuedPasskey = passkey
	return a.issueToken, a.issueExpiry, a.issueError
}

func (a *testAuthenticator) Verify(token string, _ time.Time) bool {
	a.verifyCalls = append(a.verifyCalls, token)
	return token != "" && token == a.validToken
}

type testArchiveService struct {
	beginMetadata UploadMetadata
	beginCalls    int
	entryList     []ArchiveEntry
	listCalls     int
	downloadID    string
	downloadCalls int
	deleteID      string
	deleteCalls   int
	completeID    string
	completeCalls int
}

func (s *testArchiveService) BeginUpload(_ context.Context, metadata UploadMetadata) (PresignedUpload, error) {
	s.beginCalls++
	s.beginMetadata = metadata
	return PresignedUpload{URL: "https://upload.example", Fields: map[string]string{"key": "pending/id"}}, nil
}

func (s *testArchiveService) CompleteUpload(_ context.Context, id string) (ArchiveEntry, error) {
	s.completeCalls++
	s.completeID = id
	return ArchiveEntry{ID: id, State: "available"}, nil
}

func (s *testArchiveService) List(context.Context) ([]ArchiveEntry, error) {
	s.listCalls++
	return s.entryList, nil
}

func (s *testArchiveService) Download(_ context.Context, id string) (PresignedDownload, error) {
	s.downloadCalls++
	s.downloadID = id
	return PresignedDownload{URL: "https://download.example", ExpiresAt: time.Now().Add(time.Minute)}, nil
}

func (s *testArchiveService) Delete(_ context.Context, id string) error {
	s.deleteCalls++
	s.deleteID = id
	return nil
}

func (c *testS3Client) PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return &s3.PutObjectOutput{}, nil
}

func (c *testS3Client) GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	c.getObjectCalls++
	return c.getObjectOutput, c.getObjectError
}

func (c *testS3Client) DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return &s3.DeleteObjectOutput{}, nil
}

func TestIsAllowedOriginRequiresExplicitAllowlist(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	t.Setenv("APP_ENV", "development")
	t.Setenv("AWS_LAMBDA_RUNTIME_API", "")

	for _, origin := range []string{
		"https://cv-code-vitals.github.io",
		"https://zylus08.github.io",
		"http://localhost:5173",
	} {
		if isAllowedOrigin(origin) {
			t.Fatalf("origin %q was allowed without explicit CORS configuration", origin)
		}
	}

	t.Setenv("CORS_ALLOWED_ORIGINS", "https://portal.example")
	if !isAllowedOrigin("https://portal.example") {
		t.Fatal("explicitly configured origin was rejected")
	}
}

func TestValidateCORSAllowedOriginsRequiresAtLeastOneValidOrigin(t *testing.T) {
	for _, value := range []string{"", " , , ", "javascript:alert(1)", "https://portal.example/path"} {
		if err := validateCORSAllowedOrigins(value); err == nil {
			t.Errorf("configuration %q was accepted without a valid origin", value)
		}
	}
	if err := validateCORSAllowedOrigins("https://portal.example/, http://localhost:5173"); err != nil {
		t.Fatalf("valid explicit origins were rejected: %v", err)
	}
}

func TestNewRouterKeepsHealthRouteAvailableWithoutStorageInitialization(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/status", nil)

	newRouter(Dependencies{}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected health status %d, got %d", http.StatusOK, response.Code)
	}
	var health HealthResponse
	if err := json.Unmarshal(response.Body.Bytes(), &health); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if health.Status != "SYS.INITIALIZE" {
		t.Fatalf("unexpected health status %q", health.Status)
	}
}

func TestLoginDoesNotAllowLocalDevelopmentPasskeyInProduction(t *testing.T) {
	oldPasskey := adminPasskey
	adminPasskey = ""
	t.Cleanup(func() {
		adminPasskey = oldPasskey
		sessionMu.Lock()
		sessionStore = map[string]time.Time{}
		sessionMu.Unlock()
	})
	t.Setenv("APP_ENV", "production")
	t.Setenv("LOCAL_DEV", "true")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://zylus08.github.io")

	request := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"passkey":"admin"}`))
	request.Header.Set("Origin", "https://zylus08.github.io")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	newRouter(Dependencies{}).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected production login to reject local fallback with %d, got %d", http.StatusUnauthorized, response.Code)
	}

	adminPasskey = "production-secret"
	request = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"passkey":"production-secret"}`))
	request.Header.Set("Origin", "https://zylus08.github.io")
	response = httptest.NewRecorder()
	newRouter(Dependencies{}).ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected production legacy login to be disabled without injected authenticator, got %d", response.Code)
	}
}

func TestLegacyLoginFailsClosedWhenRandomnessIsUnavailable(t *testing.T) {
	oldPasskey := adminPasskey
	adminPasskey = "local-passkey"
	oldReader := sessionTokenRandom
	sessionTokenRandom = failingRandomReader{}
	t.Cleanup(func() {
		adminPasskey = oldPasskey
		sessionTokenRandom = oldReader
		sessionMu.Lock()
		sessionStore = map[string]time.Time{}
		sessionMu.Unlock()
	})
	t.Setenv("APP_ENV", "development")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")

	request := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"passkey":"local-passkey"}`))
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	newRouter(Dependencies{}).ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("randomness failure returned %d, want 500", response.Code)
	}
	if response.Header().Get("Set-Cookie") != "" {
		t.Fatal("login issued a cookie after session-token generation failed")
	}
	if strings.Contains(response.Body.String(), "session-") {
		t.Fatal("login returned a predictable fallback session token")
	}
}

func TestLegacyLoginRotatesAndInvalidatesExistingSession(t *testing.T) {
	oldPasskey := adminPasskey
	adminPasskey = "local-passkey"
	oldSessions := sessionStore
	sessionStore = map[string]time.Time{"existing-session": time.Now().Add(time.Hour)}
	t.Cleanup(func() {
		adminPasskey = oldPasskey
		sessionStore = oldSessions
	})
	t.Setenv("APP_ENV", "development")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")

	request := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"passkey":"local-passkey"}`))
	request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "existing-session"})
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	newRouter(Dependencies{}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("login returned %d, want 200: %s", response.Code, response.Body.String())
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Value == "" || cookies[0].Value == "existing-session" {
		t.Fatalf("login did not issue a fresh session cookie: %#v", cookies)
	}
	oldRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	oldRequest.AddCookie(&http.Cookie{Name: sessionCookie, Value: "existing-session"})
	if hasValidSession(oldRequest) {
		t.Fatal("old session remained valid after successful login rotation")
	}
}

func TestProductionRoutesRejectLegacyStaticPasskeyAuthorization(t *testing.T) {
	oldPasskey := adminPasskey
	adminPasskey = "production-secret"
	t.Cleanup(func() { adminPasskey = oldPasskey })
	t.Setenv("APP_ENV", "production")

	request := httptest.NewRequest(http.MethodGet, "/api/archives", nil)
	request.Header.Set("Authorization", "Bearer production-secret")
	response := httptest.NewRecorder()
	newRouter(Dependencies{}).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected production legacy auth to reject static passkey with 401, got %d", response.Code)
	}
}

func TestInjectedAuthenticatorIssuesTokenAndAuthorizesArchiveRoutes(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://zylus08.github.io")
	auth := &testAuthenticator{
		issueToken:  "issued-bearer",
		issueExpiry: time.Now().Add(15 * time.Minute),
		validToken:  "valid-bearer",
	}
	archives := &testArchiveService{entryList: []ArchiveEntry{{ID: "opaque-id", State: "available"}}}
	handler := newRouter(Dependencies{Auth: auth, Archives: archives})

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"passkey":"injected-passkey"}`))
	loginRequest.Header.Set("Origin", "https://zylus08.github.io")
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)

	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected injected authenticator login status %d, got %d: %s", http.StatusOK, loginResponse.Code, loginResponse.Body.String())
	}
	if auth.issuedPasskey != "injected-passkey" {
		t.Fatalf("login did not pass passkey to injected authenticator: %q", auth.issuedPasskey)
	}
	var loginBody map[string]any
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &loginBody); err != nil {
		t.Fatal(err)
	}
	if loginBody["token"] != "issued-bearer" {
		t.Fatalf("login did not return injected token: %#v", loginBody)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/deliverables", nil)
	request.Header.Set("Authorization", "Bearer valid-bearer")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected injected token to authorize listing, got %d: %s", response.Code, response.Body.String())
	}
	if archives.listCalls != 1 || len(auth.verifyCalls) != 1 || auth.verifyCalls[0] != "valid-bearer" {
		t.Fatalf("injected auth/archive dependencies were not used: verify=%v listCalls=%d", auth.verifyCalls, archives.listCalls)
	}
}

func TestInjectedAuthenticatorIsRequiredForArchiveRoutes(t *testing.T) {
	auth := &testAuthenticator{validToken: "valid-bearer"}
	archives := &testArchiveService{}
	handler := newRouter(Dependencies{Auth: auth, Archives: archives})

	for _, testCase := range []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "begin upload", method: http.MethodPost, path: "/api/upload", body: `{"originalName":"report.pdf","contentType":"application/pdf","size":8,"title":"Report","version":"1","date":"2026-10-22","summary":"summary"}`},
		{name: "complete upload", method: http.MethodPost, path: "/api/upload/complete", body: `{"uploadId":"pending-id"}`},
		{name: "list", method: http.MethodGet, path: "/api/deliverables"},
		{name: "download", method: http.MethodGet, path: "/api/download?filename=archive-id"},
		{name: "delete", method: http.MethodDelete, path: "/api/archive?filename=archive-id"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(testCase.method, testCase.path, strings.NewReader(testCase.body))
			request.Header.Set("Origin", "https://zylus08.github.io")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("expected missing injected auth to be rejected, got %d", response.Code)
			}
		})
	}
	if archives.beginCalls+archives.completeCalls+archives.listCalls+archives.downloadCalls+archives.deleteCalls != 0 {
		t.Fatalf("archive service was called without injected auth: %#v", archives)
	}
}

func TestInjectedAuthenticatorProtectsEveryArchiveServiceOperation(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://zylus08.github.io")
	auth := &testAuthenticator{validToken: "valid-bearer"}
	archives := &testArchiveService{}
	handler := newRouter(Dependencies{Auth: auth, Archives: archives})
	for _, testCase := range []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodPost, path: "/api/upload", body: `{"originalName":"report.pdf","contentType":"application/pdf","size":8,"title":"Report","version":"1","date":"2026-10-22","summary":"summary"}`},
		{method: http.MethodPost, path: "/api/upload/complete", body: `{"uploadId":"pending-id"}`},
		{method: http.MethodGet, path: "/api/deliverables"},
		{method: http.MethodGet, path: "/api/download?filename=archive-id"},
		{method: http.MethodDelete, path: "/api/archive?filename=archive-id"},
	} {
		request := httptest.NewRequest(testCase.method, testCase.path, strings.NewReader(testCase.body))
		request.Header.Set("Authorization", "Bearer valid-bearer")
		request.Header.Set("Origin", "https://zylus08.github.io")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code >= 400 {
			t.Fatalf("%s %s rejected injected bearer token: %d %s", testCase.method, testCase.path, response.Code, response.Body.String())
		}
	}
	if len(auth.verifyCalls) != 5 ||
		archives.beginCalls != 1 ||
		archives.completeCalls != 1 ||
		archives.listCalls != 1 ||
		archives.downloadCalls != 1 ||
		archives.deleteCalls != 1 {
		t.Fatalf("injected auth/service dependencies were not used for each protected operation: verify=%v service=%#v", auth.verifyCalls, archives)
	}
}

func TestInjectedArchiveServiceRejectsInvalidUploadMetadataBeforeStorage(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://zylus08.github.io")
	auth := &testAuthenticator{validToken: "valid-bearer"}
	archives := &testArchiveService{}
	handler := newRouter(Dependencies{Auth: auth, Archives: archives})

	for _, testCase := range []struct {
		name    string
		payload string
	}{
		{
			name:    "unsupported zip",
			payload: `{"originalName":"archive.zip","contentType":"application/zip","size":100,"title":"Report","version":"1","date":"2026-10-22","summary":"summary"}`,
		},
		{
			name:    "invalid metadata date",
			payload: `{"originalName":"report.pdf","contentType":"application/pdf","size":100,"title":"Report","version":"1","date":"2026-02-30","summary":"summary"}`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/upload", strings.NewReader(testCase.payload))
			request.Header.Set("Authorization", "Bearer valid-bearer")
			request.Header.Set("Origin", "https://zylus08.github.io")
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code < 400 {
				t.Fatalf("expected invalid upload to fail, got %d", response.Code)
			}
		})
	}
	if archives.beginCalls != 0 {
		t.Fatalf("storage service received invalid uploads: %d", archives.beginCalls)
	}
}

func TestInjectedArchiveServiceReceivesValidatedUploadDate(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://zylus08.github.io")
	auth := &testAuthenticator{validToken: "valid-bearer"}
	archives := &testArchiveService{}
	handler := newRouter(Dependencies{Auth: auth, Archives: archives})
	const uploadDate = "2024-02-29"
	request := httptest.NewRequest(http.MethodPost, "/api/upload", strings.NewReader(
		`{"originalName":"report.pdf","contentType":"application/pdf","size":100,"title":"Report","version":"1","date":"`+uploadDate+`","summary":"summary"}`,
	))
	request.Header.Set("Authorization", "Bearer valid-bearer")
	request.Header.Set("Origin", "https://zylus08.github.io")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected valid upload status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	if archives.beginCalls != 1 {
		t.Fatalf("expected one BeginUpload call, got %d", archives.beginCalls)
	}
	if archives.beginMetadata.Date != uploadDate {
		t.Fatalf("BeginUpload received date %q, want exact date %q", archives.beginMetadata.Date, uploadDate)
	}
}
func TestLegacyUploadRejectsInvalidFilesBeforeStorage(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://zylus08.github.io")
	oldFiles := archivesFile
	oldUploads, oldUseS3, oldClient := uploadsDir, useS3, s3Client
	oldUseDB, oldPool := useDB, dbPool
	oldPasskey := adminPasskey
	t.Cleanup(func() {
		archivesFile = oldFiles
		uploadsDir, useS3, s3Client = oldUploads, oldUseS3, oldClient
		useDB, dbPool = oldUseDB, oldPool
		adminPasskey = oldPasskey
	})
	uploadsDir = t.TempDir()
	archivesFile = filepath.Join(t.TempDir(), "archives.json")
	useS3, s3Client, useDB, dbPool = false, nil, false, nil
	adminPasskey = "legacy-passkey"

	for _, testCase := range []struct {
		name        string
		contentType string
		content     []byte
	}{
		{name: "unsupported zip", contentType: "application/zip", content: arbitraryZIP(t)},
		{name: "renamed arbitrary zip", contentType: "application/vnd.openxmlformats-officedocument.presentationml.presentation", content: arbitraryZIP(t)},
		{name: "invalid pdf signature", contentType: "application/pdf", content: []byte("not a pdf")},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			body, contentType := multipartUploadBody(t, "report.pdf", testCase.contentType, testCase.content, "2026-10-22")
			if strings.Contains(testCase.name, "zip") {
				name := "archive.zip"
				if testCase.name == "renamed arbitrary zip" {
					name = "slides.pptx"
				}
				body, contentType = multipartUploadBody(t, name, testCase.contentType, testCase.content, "2026-10-22")
			}
			request := httptest.NewRequest(http.MethodPost, "/api/upload", body)
			request.Header.Set("Content-Type", contentType)
			request.Header.Set("Origin", "https://zylus08.github.io")
			request.Header.Set("Authorization", "Bearer legacy-passkey")
			response := httptest.NewRecorder()
			newRouter(Dependencies{}).ServeHTTP(response, request)

			if response.Code < 400 {
				t.Fatalf("expected invalid file to fail, got %d", response.Code)
			}
		})
	}
	entries, err := os.ReadDir(uploadsDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("invalid uploads were stored: %v", entries)
	}
}

func TestLegacyUploadUsesOpaqueUniqueStorageKeys(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://zylus08.github.io")
	oldFiles := archivesFile
	oldUploads, oldUseS3, oldClient := uploadsDir, useS3, s3Client
	oldUseDB, oldPool := useDB, dbPool
	oldPasskey := adminPasskey
	t.Cleanup(func() {
		archivesFile = oldFiles
		uploadsDir, useS3, s3Client = oldUploads, oldUseS3, oldClient
		useDB, dbPool = oldUseDB, oldPool
		adminPasskey = oldPasskey
	})
	uploadsDir = t.TempDir()
	archivesFile = filepath.Join(t.TempDir(), "archives.json")
	useS3, s3Client, useDB, dbPool = false, nil, false, nil
	adminPasskey = "legacy-passkey"

	var keys []string
	for i := 0; i < 2; i++ {
		body, contentType := multipartUploadBody(t, "same-report.pdf", "application/pdf", []byte("%PDF-1.7 valid"), "2026-10-22")
		request := httptest.NewRequest(http.MethodPost, "/api/upload", body)
		request.Header.Set("Content-Type", contentType)
		request.Header.Set("Origin", "https://zylus08.github.io")
		request.Header.Set("Authorization", "Bearer legacy-passkey")
		response := httptest.NewRecorder()
		newRouter(Dependencies{}).ServeHTTP(response, request)
		if response.Code != http.StatusCreated {
			t.Fatalf("expected upload status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
		}
		var entry ArchiveEntry
		if err := json.Unmarshal(response.Body.Bytes(), &entry); err != nil {
			t.Fatal(err)
		}
		keys = append(keys, entry.Filename)
		if strings.Contains(entry.Filename, "same-report") || entry.OriginalName != "same-report.pdf" {
			t.Fatalf("storage key or display metadata was not isolated from user filename: %#v", entry)
		}
	}
	if keys[0] == keys[1] {
		t.Fatalf("same-name uploads reused storage key %q", keys[0])
	}
}

func multipartUploadBody(t *testing.T, filename, contentType string, contents []byte, date string) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(contents); err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]string{
		"title": "Report", "version": "1", "date": date, "summary": "summary",
	} {
		if err := writer.WriteField(name, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return &body, writer.FormDataContentType()
}

func TestProductionConfigurationRequiresDurableAWSStorage(t *testing.T) {
	err := validateProductionConfig("production", "", "us-east-1", "postgres://db")
	if err == nil || !strings.Contains(err.Error(), "S3_BUCKET") {
		t.Fatalf("expected missing S3_BUCKET error, got %v", err)
	}

	err = validateProductionConfig("production", "archive-bucket", "", "postgres://db")
	if err == nil || !strings.Contains(err.Error(), "S3_REGION") {
		t.Fatalf("expected missing S3_REGION error, got %v", err)
	}

	err = validateProductionConfig("production", "archive-bucket", "us-east-1", "")
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected missing DATABASE_URL error, got %v", err)
	}

	if err := validateProductionConfig("production", "archive-bucket", "us-east-1", "postgres://db"); err != nil {
		t.Fatalf("expected complete production configuration to be accepted, got %v", err)
	}
	if err := validateProductionConfig("", "", "", ""); err != nil {
		t.Fatalf("expected local development configuration to remain optional, got %v", err)
	}
}

func TestDownloadHandlerStreamsAuthenticatedPrivateS3Object(t *testing.T) {
	oldUseS3, oldBucket, oldClient := useS3, s3Bucket, s3Client
	oldArchivesFile, oldUseDB, oldPool := archivesFile, useDB, dbPool
	oldUploadsDir, oldPasskey := uploadsDir, adminPasskey
	t.Cleanup(func() {
		useS3, s3Bucket, s3Client = oldUseS3, oldBucket, oldClient
		archivesFile, useDB, dbPool = oldArchivesFile, oldUseDB, oldPool
		uploadsDir, adminPasskey = oldUploadsDir, oldPasskey
		sessionMu.Lock()
		sessionStore = map[string]time.Time{}
		sessionMu.Unlock()
	})

	useS3 = true
	s3Bucket = "private-archive"
	client := &testS3Client{
		getObjectOutput: &s3.GetObjectOutput{
			Body:        io.NopCloser(bytes.NewReader([]byte("private contents"))),
			ContentType: stringPointer("application/pdf"),
		},
	}
	s3Client = client
	useDB = false
	dbPool = nil
	adminPasskey = "test-passkey"
	archivesFile = filepath.Join(t.TempDir(), "archives.json")
	uploadsDir = t.TempDir()
	entries, err := json.Marshal([]ArchiveEntry{{
		Filename:     "stored-report.pdf",
		OriginalName: "Report.pdf",
		URL:          "/api/download?filename=stored-report.pdf",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(archivesFile, entries, 0o600); err != nil {
		t.Fatal(err)
	}

	sessionMu.Lock()
	sessionStore = map[string]time.Time{"valid-session": time.Now().Add(time.Hour)}
	sessionMu.Unlock()

	request := httptest.NewRequest(http.MethodGet, "/download?filename=stored-report.pdf", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "valid-session"})
	response := httptest.NewRecorder()

	downloadHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if got := response.Body.String(); got != "private contents" {
		t.Fatalf("expected streamed object contents, got %q", got)
	}
	if got := response.Header().Get("Content-Type"); got != "application/pdf" {
		t.Fatalf("expected S3 content type, got %q", got)
	}
	if got := response.Header().Get("Content-Disposition"); !strings.Contains(got, "attachment") || !strings.Contains(got, "Report.pdf") {
		t.Fatalf("expected attachment filename header, got %q", got)
	}
	if got := response.Header().Get("Cache-Control"); got != "private, no-store" {
		t.Fatalf("expected private no-store caching policy, got %q", got)
	}
	if client.getObjectCalls != 1 {
		t.Fatalf("expected one private S3 read, got %d", client.getObjectCalls)
	}
}

func TestDownloadHandlerSupportsLocalDevelopmentFiles(t *testing.T) {
	oldUseS3, oldClient := useS3, s3Client
	oldArchivesFile, oldUseDB, oldPool := archivesFile, useDB, dbPool
	oldUploadsDir, oldPasskey := uploadsDir, adminPasskey
	t.Cleanup(func() {
		useS3, s3Client = oldUseS3, oldClient
		archivesFile, useDB, dbPool = oldArchivesFile, oldUseDB, oldPool
		uploadsDir, adminPasskey = oldUploadsDir, oldPasskey
		sessionMu.Lock()
		sessionStore = map[string]time.Time{}
		sessionMu.Unlock()
	})

	useS3, s3Client = false, nil
	useDB, dbPool = false, nil
	adminPasskey = "test-passkey"
	uploadsDir = t.TempDir()
	archivesFile = filepath.Join(t.TempDir(), "archives.json")
	if err := os.WriteFile(filepath.Join(uploadsDir, "local-report.pdf"), []byte("local contents"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(archivesFile, []byte(`[{"filename":"local-report.pdf","originalName":"Local Report.pdf"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	sessionMu.Lock()
	sessionStore = map[string]time.Time{"valid-session": time.Now().Add(time.Hour)}
	sessionMu.Unlock()

	request := httptest.NewRequest(http.MethodGet, "/download?filename=local-report.pdf", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "valid-session"})
	response := httptest.NewRecorder()
	downloadHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if got := response.Body.String(); got != "local contents" {
		t.Fatalf("expected local file contents, got %q", got)
	}
}

func TestDownloadHandlerRequiresAuthenticationBeforeReadingS3(t *testing.T) {
	oldUseS3, oldClient, oldArchivesFile, oldUseDB, oldPool := useS3, s3Client, archivesFile, useDB, dbPool
	oldPasskey := adminPasskey
	t.Cleanup(func() {
		useS3, s3Client, archivesFile, useDB, dbPool = oldUseS3, oldClient, oldArchivesFile, oldUseDB, oldPool
		adminPasskey = oldPasskey
		sessionMu.Lock()
		sessionStore = map[string]time.Time{}
		sessionMu.Unlock()
	})

	client := &testS3Client{getObjectOutput: &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(nil))}}
	useS3, s3Client, useDB, dbPool = true, client, false, nil
	adminPasskey = "test-passkey"
	archivesFile = filepath.Join(t.TempDir(), "archives.json")
	if err := os.WriteFile(archivesFile, []byte(`[{"filename":"file.pdf"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	sessionMu.Lock()
	sessionStore = map[string]time.Time{}
	sessionMu.Unlock()

	request := httptest.NewRequest(http.MethodGet, "/download?filename=file.pdf", nil)
	response := httptest.NewRecorder()
	downloadHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status, got %d", response.Code)
	}
	if client.getObjectCalls != 0 {
		t.Fatalf("expected no S3 read for unauthorized request, got %d", client.getObjectCalls)
	}
}

func TestDownloadHandlerRejectsUnknownAndInvalidArchiveKeys(t *testing.T) {
	oldUseS3, oldClient, oldArchivesFile, oldUseDB, oldPool := useS3, s3Client, archivesFile, useDB, dbPool
	oldPasskey := adminPasskey
	t.Cleanup(func() {
		useS3, s3Client, archivesFile, useDB, dbPool = oldUseS3, oldClient, oldArchivesFile, oldUseDB, oldPool
		adminPasskey = oldPasskey
		sessionMu.Lock()
		sessionStore = map[string]time.Time{}
		sessionMu.Unlock()
	})

	client := &testS3Client{getObjectOutput: &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(nil))}}
	useS3, s3Client, useDB, dbPool = true, client, false, nil
	adminPasskey = "test-passkey"
	archivesFile = filepath.Join(t.TempDir(), "archives.json")
	if err := os.WriteFile(archivesFile, []byte(`[{"filename":"file.pdf"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	sessionMu.Lock()
	sessionStore = map[string]time.Time{"valid-session": time.Now().Add(time.Hour)}
	sessionMu.Unlock()

	for _, testCase := range []struct {
		name     string
		filename string
		wantCode int
	}{
		{name: "unknown archive", filename: "other.pdf", wantCode: http.StatusNotFound},
		{name: "path traversal", filename: "../file.pdf", wantCode: http.StatusBadRequest},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/download?filename="+testCase.filename, nil)
			request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "valid-session"})
			response := httptest.NewRecorder()
			downloadHandler(response, request)

			if response.Code != testCase.wantCode {
				t.Fatalf("expected status %d, got %d", testCase.wantCode, response.Code)
			}
		})
	}
	if client.getObjectCalls != 0 {
		t.Fatalf("expected invalid or unknown keys not to access S3, got %d calls", client.getObjectCalls)
	}
}

func TestDownloadHandlerReportsS3Failures(t *testing.T) {
	oldUseS3, oldClient, oldArchivesFile, oldUseDB, oldPool := useS3, s3Client, archivesFile, useDB, dbPool
	oldPasskey := adminPasskey
	t.Cleanup(func() {
		useS3, s3Client, archivesFile, useDB, dbPool = oldUseS3, oldClient, oldArchivesFile, oldUseDB, oldPool
		adminPasskey = oldPasskey
		sessionMu.Lock()
		sessionStore = map[string]time.Time{}
		sessionMu.Unlock()
	})

	client := &testS3Client{getObjectError: errors.New("S3 unavailable")}
	useS3, s3Client, useDB, dbPool = true, client, false, nil
	adminPasskey = "test-passkey"
	archivesFile = filepath.Join(t.TempDir(), "archives.json")
	if err := os.WriteFile(archivesFile, []byte(`[{"filename":"file.pdf","originalName":"File.pdf"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	sessionMu.Lock()
	sessionStore = map[string]time.Time{"valid-session": time.Now().Add(time.Hour)}
	sessionMu.Unlock()

	request := httptest.NewRequest(http.MethodGet, "/download?filename=file.pdf", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "valid-session"})
	response := httptest.NewRecorder()
	downloadHandler(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("expected storage failure status %d, got %d", http.StatusBadGateway, response.Code)
	}
}

func stringPointer(value string) *string {
	return &value
}
