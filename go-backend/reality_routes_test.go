package main

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/execution"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRealityLiveRoutesRejectAbsentSourceAndDemo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	setupRemediationRoutes(r, func(c *gin.Context) { c.Next() }, nil, execution.NewUnavailableContainerRunner())
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{"GET", "/api/remediation/targets", "", 400},
		{"GET", "/api/remediation/targets?workspace_id=wrong", "", 422},
		{"POST", "/api/remediation/analyze", `{"target_id":"INJ-EXT2-RES-02"}`, 400},
		{"GET", "/api/remediation/demo/targets", "", 200},
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != tc.status {
			t.Errorf("%s: got %d %s", tc.path, w.Code, w.Body.String())
		}
		if tc.status != http.StatusOK && strings.Contains(w.Body.String(), `"targets"`) {
			t.Error("failed live request exposed targets")
		}
	}
}
