package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Prof-bilal/ShieldCI/backend/internal/runner"
	"github.com/Prof-bilal/ShieldCI/backend/internal/store"
	"github.com/Prof-bilal/ShieldCI/backend/internal/wardenadapter"
)

func TestAPIRunnerAuditFlow(t *testing.T) {
	db := store.NewMemory()
	mux := newMux(server{store: db})
	policyBody := `{"version":"v1","name":"e2e-fixture","filesystem":{"read":[],"write":[]},"network":{"allow":[]},"environment":{"allow":[]}}`
	policyResponse := httptest.NewRecorder()
	mux.ServeHTTP(policyResponse, httptest.NewRequest(http.MethodPost, "/api/v1/policies", strings.NewReader(policyBody)))
	if policyResponse.Code != http.StatusCreated {
		t.Fatalf("policy create failed: %d %s", policyResponse.Code, policyResponse.Body.String())
	}
	var policyCreated struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(policyResponse.Body.Bytes(), &policyCreated); err != nil {
		t.Fatal(err)
	}
	jobBody := `{"policy_id":"` + policyCreated.ID + `","command":["/bin/echo","e2e"]}`
	jobResponse := httptest.NewRecorder()
	mux.ServeHTTP(jobResponse, httptest.NewRequest(http.MethodPost, "/api/v1/jobs", strings.NewReader(jobBody)))
	if jobResponse.Code != http.StatusAccepted {
		t.Fatalf("job create failed: %d %s", jobResponse.Code, jobResponse.Body.String())
	}
	var jobCreated struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(jobResponse.Body.Bytes(), &jobCreated); err != nil {
		t.Fatal(err)
	}

	fake := filepath.Join(t.TempDir(), "warden")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	adapter, err := wardenadapter.New(wardenadapter.Config{BinaryPath: fake, TempDir: t.TempDir(), WorkingDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if claimed, err := (runner.Runner{Store: db, Adapter: adapter}).RunNext(context.Background()); err != nil || !claimed {
		t.Fatalf("runner failed: claimed=%v err=%v", claimed, err)
	}

	statusResponse := httptest.NewRecorder()
	mux.ServeHTTP(statusResponse, httptest.NewRequest(http.MethodGet, "/api/v1/jobs/"+jobCreated.ID, nil))
	if statusResponse.Code != http.StatusOK || !strings.Contains(statusResponse.Body.String(), `"status":"passed"`) {
		t.Fatalf("unexpected job status: %d %s", statusResponse.Code, statusResponse.Body.String())
	}
	auditResponse := httptest.NewRecorder()
	mux.ServeHTTP(auditResponse, httptest.NewRequest(http.MethodGet, "/api/v1/jobs/"+jobCreated.ID+"/audit", nil))
	if auditResponse.Code != http.StatusOK || !strings.Contains(auditResponse.Body.String(), `"decision":"allowed"`) {
		t.Fatalf("unexpected audit response: %d %s", auditResponse.Code, auditResponse.Body.String())
	}
}
