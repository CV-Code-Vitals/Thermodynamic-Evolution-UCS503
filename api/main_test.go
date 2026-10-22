package main

import "testing"

func TestIsAllowedOriginAllowsCurrentGitHubPagesSite(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	if !isAllowedOrigin("https://cv-code-vitals.github.io") {
		t.Fatal("expected the published Admin Portal origin to be allowed")
	}
}
