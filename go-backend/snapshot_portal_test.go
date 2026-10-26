package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saksham/self-healing-pipeline/agent"
)

func TestSnapshotPortalScanEvolutionAndPatch(t *testing.T) {
	engine := os.Getenv("ENGINE_PATH")
	if engine == "" {
		t.Skip("set ENGINE_PATH to run the real snapshot portal integration")
	}
	t.Setenv("SANDBOX_SERVICE_URL", "")
	t.Setenv("MUTATION_API_KEY", "")
	entries := map[string]string{}
	for _, name := range []string{"go.mod", "benchmark_target.go", "benchmark_target_test.go"} {
		content, err := os.ReadFile(filepath.Join("..", "samples", "benchmark", name))
		if err != nil {
			t.Fatal(err)
		}
		entries["repo/samples/benchmark/"+name] = string(content)
	}
	tempRoot := t.TempDir()
	store := NewSnapshotStore(&fakeSnapshotArchiveReader{body: makeSnapshotArchive(t, entries)}, tempRoot)
	scan := snapshotScanRepoHandler(engine, time.Minute, store, NewPublicScanController(time.Minute, 1))
	response := serveScanRequest(scan, `{"snapshot_id":"`+testSnapshotID+`","retain_workspace":true}`, "8.8.8.8", "")
	if response.Code != http.StatusOK {
		t.Fatalf("scan: %d %s", response.Code, response.Body.String())
	}
	var result struct {
		Report      ThermodynamicReport `json:"report"`
		WorkspaceID string              `json:"workspace_id"`
		Targets     []evolutionTarget   `json:"targets"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.WorkspaceID != testSnapshotID || result.Report.FilesAnalyzed == 0 || len(result.Targets) != 1 {
		t.Fatalf("snapshot scan did not provide the portal contract: %+v", result)
	}
	target := result.Targets[0]
	request := EvolveRequest{SnapshotID: testSnapshotID, FilePath: target.FilePath, FunctionName: target.FunctionName,
		StartByte: target.StartByte, SourceHash: target.SourceHash, Config: agent.AnnealingConfig{MaxSteps: 4, InitialTemp: 100, CoolingRate: 0.85}}
	invoke := func(payload EvolveRequest) *httptest.ResponseRecorder {
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		router := gin.New()
		router.POST("/evolve", evolveHandler(newWorkspaceStore(), store))
		response := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/evolve", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, req)
		return response
	}
	stale := request
	stale.SourceHash = strings.Repeat("0", 64)
	if response := invoke(stale); response.Code != http.StatusBadRequest {
		t.Fatalf("stale snapshot accepted: %s", response.Body.String())
	}
	response = invoke(request)
	if response.Code != http.StatusOK {
		t.Fatalf("evolve: %d %s", response.Code, response.Body.String())
	}
	marker := "event: complete\ndata: "
	_, data, ok := strings.Cut(response.Body.String(), marker)
	if !ok {
		t.Fatalf("no completion: %s", response.Body.String())
	}
	var evolved agent.AnnealingResult
	if err := json.Unmarshal([]byte(strings.SplitN(data, "\n", 2)[0]), &evolved); err != nil {
		t.Fatal(err)
	}
	if !evolved.BestImproved || evolved.CorrectnessPassed == 0 || evolved.Patch == "" {
		t.Fatalf("snapshot did not export a tested improvement: %+v", evolved)
	}
	patchRoot := t.TempDir()
	for name, content := range entries {
		path := filepath.Join(patchRoot, strings.TrimPrefix(name, "repo/"))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.Command("git", "apply", "--check", "-")
	command.Dir = patchRoot
	command.Stdin = strings.NewReader(evolved.Patch)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("exported snapshot patch does not apply: %v %s\n%s", err, output, evolved.Patch)
	}
	remaining, err := os.ReadDir(tempRoot)
	if err != nil || len(remaining) != 0 {
		t.Fatalf("snapshot workspaces leaked: %v %v", remaining, err)
	}
}
