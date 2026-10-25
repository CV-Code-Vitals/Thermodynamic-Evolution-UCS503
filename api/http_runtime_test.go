package main

import (
	"context"
	"strings"
	"testing"
)

func TestProductionHTTPRequiresExplicitRuntimeAndAWSStorage(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("API_RUNTIME", "")
	if _, err := newHTTPDependencies(context.Background()); err == nil || !strings.Contains(err.Error(), "API_RUNTIME=http") {
		t.Fatalf("production HTTP started without explicit opt-in: %v", err)
	}
	t.Setenv("API_RUNTIME", "http")
	t.Setenv("AWS_REGION", "")
	t.Setenv("S3_BUCKET", "")
	t.Setenv("DYNAMODB_TABLE", "")
	t.Setenv("ADMIN_PASSKEY_PARAMETER", "")
	t.Setenv("TOKEN_HMAC_KEY_PARAMETER", "")
	deps, err := newHTTPDependencies(context.Background())
	if err == nil || deps.Archives != nil || deps.LocalFiles != nil {
		t.Fatalf("production HTTP fell back to local storage: %+v, %v", deps, err)
	}
}
