package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
)

func TestReadLambdaConfigRequiresExplicitCORSOrigins(t *testing.T) {
	values := map[string]string{
		"AWS_REGION":               "ap-south-1",
		"S3_BUCKET":                "archive-bucket",
		"DYNAMODB_TABLE":           "archive-table",
		"ADMIN_PASSKEY_PARAMETER":  "/app/admin/passkey",
		"TOKEN_HMAC_KEY_PARAMETER": "/app/admin/token-key",
	}
	if _, err := readLambdaConfig(func(name string) string { return values[name] }); err == nil {
		t.Fatal("Lambda configuration succeeded without an explicit CORS allowlist")
	}
	values["CORS_ALLOWED_ORIGINS"] = " , "
	if _, err := readLambdaConfig(func(name string) string { return values[name] }); err == nil {
		t.Fatal("Lambda configuration accepted an empty CORS allowlist")
	}
}

type lambdaContractArchiveService struct {
	listCalls     int
	downloadCalls int
	listDeadline  bool
	downloadError error
}

func (s *lambdaContractArchiveService) BeginUpload(context.Context, UploadMetadata) (PresignedUpload, error) {
	return PresignedUpload{}, nil
}

func (s *lambdaContractArchiveService) CompleteUpload(context.Context, string) (ArchiveEntry, error) {
	return ArchiveEntry{}, nil
}

func (s *lambdaContractArchiveService) List(ctx context.Context) ([]ArchiveEntry, error) {
	s.listCalls++
	_, s.listDeadline = ctx.Deadline()
	return nil, nil
}

func (s *lambdaContractArchiveService) Download(context.Context, string) (PresignedDownload, error) {
	s.downloadCalls++
	if s.downloadError != nil {
		return PresignedDownload{}, s.downloadError
	}
	return PresignedDownload{}, ErrArchiveNotFound
}

func (s *lambdaContractArchiveService) Delete(context.Context, string) error {
	return ErrArchiveNotFound
}

func TestLambdaSharedRouterRejectsUnauthenticatedArchiveRoutes(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	auth := &testAuthenticator{validToken: "valid-token"}
	archives := &lambdaContractArchiveService{}
	handler := newRouter(Dependencies{Auth: auth, Archives: archives})

	for _, testCase := range []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/api/upload", `{"originalName":"report.pdf","contentType":"application/pdf","size":8,"title":"Report","version":"1","date":"2026-10-22","summary":"summary"}`},
		{http.MethodPost, "/api/upload/complete", `{"uploadId":"pending-id"}`},
		{http.MethodGet, "/api/deliverables", ""},
		{http.MethodGet, "/api/download?filename=archive-id", ""},
		{http.MethodDelete, "/api/archive?filename=archive-id", ""},
	} {
		request := httptest.NewRequest(testCase.method, testCase.path, strings.NewReader(testCase.body))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Errorf("%s %s returned %d, want 401", testCase.method, testCase.path, response.Code)
		}
	}
	if archives.listCalls != 0 || archives.downloadCalls != 0 {
		t.Fatalf("unauthorized requests reached archive service: %#v", archives)
	}
}

func TestLambdaSharedRouterMapsUnknownArchivesToNotFound(t *testing.T) {
	auth := &testAuthenticator{validToken: "valid-token"}
	handler := newRouter(Dependencies{Auth: auth, Archives: &lambdaContractArchiveService{}})
	request := httptest.NewRequest(http.MethodGet, "/api/download?filename=missing-id", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("unknown archive returned %d, want 404: %s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("error content type = %q, want application/json", got)
	}
	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode stable error response: %v", err)
	}
	if body["error"] != "archive_not_found" {
		t.Fatalf("error code = %q, want archive_not_found", body["error"])
	}
}

func TestLambdaSharedRouterRejectsMalformedOrExpiredTokenBeforeStorage(t *testing.T) {
	auth := &testAuthenticator{}
	archives := &lambdaContractArchiveService{}
	handler := newRouter(Dependencies{Auth: auth, Archives: archives})
	for _, token := range []string{"malformed", "well-formed-but-expired"} {
		request := httptest.NewRequest(http.MethodGet, "/api/deliverables", nil)
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized || archives.listCalls != 0 {
			t.Fatalf("invalid token returned %d and reached storage %d times", response.Code, archives.listCalls)
		}
	}
}

func TestLambdaSharedRouterRequiresTrustedOriginForMutationsAndRejectsMethods(t *testing.T) {
	archives := &lambdaContractArchiveService{}
	handler := newRouter(Dependencies{
		Auth:     &testAuthenticator{validToken: "valid-token"},
		Archives: archives,
	})

	request := httptest.NewRequest(http.MethodDelete, "/api/archive?filename=archive-id", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("mutation without trusted origin returned %d, want 403", response.Code)
	}

	request = httptest.NewRequest(http.MethodPut, "/api/deliverables", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unsupported method returned %d, want 405", response.Code)
	}
}

func TestProductionMutationOriginsExcludeLocalhost(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("AWS_LAMBDA_RUNTIME_API", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://cv-code-vitals.github.io")

	if !isAllowedOrigin("https://cv-code-vitals.github.io") {
		t.Fatal("production configuration rejected its configured Admin Portal origin")
	}
	if isAllowedOrigin("https://zylus08.github.io") {
		t.Fatal("production configuration allowed an origin not in its allow-list")
	}
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	t.Setenv("AWS_LAMBDA_RUNTIME_API", "127.0.0.1:9001")
	if isAllowedOrigin("https://zylus08.github.io") {
		t.Fatal("Lambda runtime allowed a hard-coded production origin without configuration")
	}
	if isAllowedOrigin("http://localhost:5173") {
		t.Fatal("Lambda runtime trusted a local development origin without APP_ENV")
	}
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://custom-admin.example")
	if !isAllowedOrigin("https://custom-admin.example") {
		t.Fatal("Lambda runtime rejected an explicitly configured production origin")
	}
}

func TestLambdaArchiveCallsUseRequestScopedTimeout(t *testing.T) {
	archives := &lambdaContractArchiveService{}
	handler := newRouter(Dependencies{
		Auth:     &testAuthenticator{validToken: "valid-token"},
		Archives: archives,
	})
	request := httptest.NewRequest(http.MethodGet, "/api/deliverables", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("list returned %d: %s", response.Code, response.Body.String())
	}
	if !archives.listDeadline {
		t.Fatal("archive service did not receive a request-scoped deadline")
	}
}

func TestLambdaErrorLoggingDoesNotExposeSecretOrPresignedURL(t *testing.T) {
	const secret = "do-not-log-this"
	const signedURL = "https://private.example/file?signature=secret"
	const bearer = "valid-token"
	const internalToken = "authorizer-context-token-that-must-not-be-logged"
	loggerOutput := &strings.Builder{}
	logger := slog.New(slog.NewJSONHandler(loggerOutput, nil))
	archives := &lambdaContractArchiveService{downloadError: errors.New(secret + " " + signedURL + " " + internalToken)}
	handler := newRouter(Dependencies{
		Auth:     &testAuthenticator{validToken: "valid-token"},
		Archives: archives,
		Logger:   logger,
	})
	request := httptest.NewRequest(http.MethodGet, "/api/download?filename=missing-id", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	request.Header.Set("X-Test-Secret", secret+signedURL)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway || loggerOutput.Len() == 0 {
		t.Fatalf("expected structured internal-error log, status=%d logs=%q", response.Code, loggerOutput.String())
	}
	for _, sensitiveValue := range []string{secret, signedURL, bearer, internalToken, "Authorization"} {
		if strings.Contains(loggerOutput.String(), sensitiveValue) {
			t.Fatalf("logs exposed %q: %s", sensitiveValue, loggerOutput.String())
		}
	}
}

func TestAPIGatewayV2HandlerMapsRequestAndResponse(t *testing.T) {
	handler := newAPIGatewayV2Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lambdaContext, ok := lambdacontext.FromContext(r.Context())
		if !ok || lambdaContext.AwsRequestID != "request-id" {
			http.Error(w, "Lambda request context missing", http.StatusInternalServerError)
			return
		}
		gatewayContext, ok := core.GetAPIGatewayV2ContextFromContext(r.Context())
		if !ok || gatewayContext.APIID != "api-id" || gatewayContext.RequestID != "request-id" {
			http.Error(w, "API Gateway request context missing", http.StatusInternalServerError)
			return
		}
		body, err := json.Marshal(map[string]string{
			"method":     r.Method,
			"path":       r.URL.Path,
			"query":      r.URL.Query().Get("search"),
			"testHeader": r.Header.Get("X-Request-Test"),
			"body":       readRequestBody(r),
			"apiID":      gatewayContext.APIID,
		})
		if err != nil {
			http.Error(w, "encode request", http.StatusInternalServerError)
			return
		}
		w.Header().Set("X-Response-Test", "mapped")
		w.Header().Add("Set-Cookie", "first=value; Path=/")
		w.Header().Add("Set-Cookie", "second=value; Path=/")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
	}))

	lambdaCtx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "request-id",
	})
	response, err := handler(lambdaCtx, events.APIGatewayV2HTTPRequest{
		Version:        "2.0",
		RouteKey:       "POST /api/items",
		RawPath:        "/api/items",
		RawQueryString: "search=one%20two",
		Headers: map[string]string{
			"content-type":   "application/json",
			"x-request-test": "header-value",
		},
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			APIID:     "api-id",
			RequestID: "request-id",
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "POST",
				Path:   "/api/items",
			},
		},
		Body: `{"item":"sample"}`,
	})
	if err != nil {
		t.Fatalf("Lambda adapter returned an error: %v", err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", response.StatusCode)
	}
	if response.Headers["X-Response-Test"] != "mapped" {
		t.Fatalf("response headers were not mapped: %#v", response.Headers)
	}
	if len(response.Cookies) != 2 {
		t.Fatalf("multiple response cookies were not preserved: %#v", response.Cookies)
	}
	var mapped map[string]string
	if err := json.Unmarshal([]byte(response.Body), &mapped); err != nil {
		t.Fatalf("decode mapped response: %v", err)
	}
	if mapped["method"] != "POST" ||
		mapped["path"] != "/api/items" ||
		mapped["query"] != "one two" ||
		mapped["testHeader"] != "header-value" ||
		mapped["body"] != `{"item":"sample"}` ||
		mapped["apiID"] != "api-id" {
		t.Fatalf("API Gateway request fields were not preserved: %#v", mapped)
	}
}

func readRequestBody(r *http.Request) string {
	body, _ := io.ReadAll(r.Body)
	return string(body)
}

type testSSMClient struct {
	output   *ssm.GetParameterOutput
	metadata *ssm.DescribeParametersOutput
	err      error
}

func (c testSSMClient) GetParameter(context.Context, *ssm.GetParameterInput, ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	return c.output, c.err
}

func (c testSSMClient) DescribeParameters(context.Context, *ssm.DescribeParametersInput, ...func(*ssm.Options)) (*ssm.DescribeParametersOutput, error) {
	return c.metadata, c.err
}

func TestLoadSSMSecureStringRequiresStandardTierAndAWSManagedKey(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		metadata  *types.ParameterMetadata
		wantError bool
	}{
		{
			name: "valid standard secure string",
			metadata: &types.ParameterMetadata{
				Name:  aws.String("/admin/secret"),
				Type:  types.ParameterTypeSecureString,
				Tier:  types.ParameterTierStandard,
				KeyId: aws.String("alias/aws/ssm"),
			},
		},
		{
			name: "plain string",
			metadata: &types.ParameterMetadata{
				Name: aws.String("/admin/secret"),
				Type: types.ParameterTypeString,
				Tier: types.ParameterTierStandard,
			},
			wantError: true,
		},
		{
			name: "advanced tier",
			metadata: &types.ParameterMetadata{
				Name:  aws.String("/admin/secret"),
				Type:  types.ParameterTypeSecureString,
				Tier:  types.ParameterTierAdvanced,
				KeyId: aws.String("alias/aws/ssm"),
			},
			wantError: true,
		},
		{
			name: "customer managed key",
			metadata: &types.ParameterMetadata{
				Name:  aws.String("/admin/secret"),
				Type:  types.ParameterTypeSecureString,
				Tier:  types.ParameterTierStandard,
				KeyId: aws.String("arn:aws:kms:us-east-1:123456789012:key/custom"),
			},
			wantError: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			value, err := loadSSMSecureString(context.Background(), testSSMClient{
				output: &ssm.GetParameterOutput{Parameter: &types.Parameter{
					Name:  aws.String("/admin/secret"),
					Type:  types.ParameterTypeSecureString,
					Value: aws.String("secret-value"),
				}},
				metadata: &ssm.DescribeParametersOutput{
					Parameters: []types.ParameterMetadata{*testCase.metadata},
				},
			}, "/admin/secret")
			if testCase.wantError {
				if err == nil {
					t.Fatalf("accepted invalid SecureString configuration with value %q", value)
				}
				return
			}
			if err != nil {
				t.Fatalf("valid SecureString rejected: %v", err)
			}
			if value != "secret-value" {
				t.Fatalf("loaded value = %q, want secret-value", value)
			}
		})
	}
}

func TestNewLocalAuthenticatorUsesConfiguredOrEphemeralHMACKey(t *testing.T) {
	first, err := newLocalAuthenticator("local-passkey", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := newLocalAuthenticator("local-passkey", "")
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := first.Issue("local-passkey")
	if err != nil {
		t.Fatal(err)
	}
	if !first.Verify(token, time.Now()) {
		t.Fatal("local authenticator did not verify its own token")
	}
	if second.Verify(token, time.Now()) {
		t.Fatal("ephemeral signing key was shared across local authenticator instances")
	}

	configured, err := newLocalAuthenticator("local-passkey", strings.Repeat("k", 32))
	if err != nil {
		t.Fatal(err)
	}
	token, _, err = configured.Issue("local-passkey")
	if err != nil || !configured.Verify(token, time.Now()) {
		t.Fatalf("configured local signing key failed: %v", err)
	}
	if _, err := newLocalAuthenticator("", ""); err == nil {
		t.Fatal("accepted an empty local passkey")
	}
}

func TestNewLocalArchiveDependenciesDoNotRequireAWSConfiguration(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_LAMBDA_RUNTIME_API", "")
	temp := t.TempDir()
	deps, err := newLocalDependencies("local-passkey", "", filepath.Join(temp, "archives.json"), filepath.Join(temp, "uploads"))
	if err != nil {
		t.Fatalf("local startup required AWS configuration: %v", err)
	}
	token, _, err := deps.Auth.Issue("local-passkey")
	if err != nil || !deps.Auth.Verify(token, time.Now()) {
		t.Fatalf("local dependencies did not issue a working bearer token: %v", err)
	}
	if deps.Archives == nil || deps.LocalFiles == nil {
		t.Fatal("local startup did not wire archive service and file store")
	}
}
