package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Prof-bilal/ShieldCI/backend/internal/store"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	(server{}).health(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("unexpected health response: %d %s", rec.Code, rec.Body.String())
	}
}

func TestCreatePolicyAndQueueJob(t *testing.T) {
	svc := server{store: store.NewMemory()}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/policies", svc.policies)
	mux.HandleFunc("/api/v1/policies/", svc.policies)
	mux.HandleFunc("/api/v1/jobs", svc.jobs)
	mux.HandleFunc("/api/v1/jobs/", svc.jobs)

	policyBody := `{"version":"v1","name":"fixture","filesystem":{"read":["./workspace"]},"network":{"allow":[]},"environment":{"allow":[]}}`
	createPolicy := httptest.NewRequest(http.MethodPost, "/api/v1/policies", strings.NewReader(policyBody))
	policyResponse := httptest.NewRecorder()
	mux.ServeHTTP(policyResponse, createPolicy)
	if policyResponse.Code != http.StatusCreated || !strings.Contains(policyResponse.Body.String(), `"id":"pol_`) {
		t.Fatalf("unexpected policy response: %d %s", policyResponse.Code, policyResponse.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(policyResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	jobBody := `{"policy_id":"` + created.ID + `","command":["/bin/echo","hello"]}`
	createJob := httptest.NewRequest(http.MethodPost, "/api/v1/jobs", strings.NewReader(jobBody))
	jobResponse := httptest.NewRecorder()
	mux.ServeHTTP(jobResponse, createJob)
	if jobResponse.Code != http.StatusAccepted || !strings.Contains(jobResponse.Body.String(), `"status":"queued"`) {
		t.Fatalf("unexpected job response: %d %s", jobResponse.Code, jobResponse.Body.String())
	}
}

func TestSecretScanStoresRedactedFinding(t *testing.T) {
	db := store.NewMemory()
	svc := server{store: db}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "config.env"), []byte("AWS_KEY=AKIA1234567890ABCDEF\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/scans/secrets", strings.NewReader(`{"root":"`+root+`"}`))
	rec := httptest.NewRecorder()
	svc.findings(rec, req)
	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), `"count":1`) || strings.Contains(rec.Body.String(), "AKIA") {
		t.Fatalf("unexpected scan response: %d %s", rec.Code, rec.Body.String())
	}
	list := httptest.NewRecorder()
	svc.findings(list, httptest.NewRequest(http.MethodGet, "/api/v1/findings", nil))
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"severity":"high"`) {
		t.Fatalf("unexpected findings response: %d %s", list.Code, list.Body.String())
	}
}

func TestAuthMiddlewareProtectsAPIWhenConfigured(t *testing.T) {
	handler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }), "test-token")
	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/jobs", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", unauthorized.Code)
	}
	authorized := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/jobs", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	handler.ServeHTTP(authorized, request)
	if authorized.Code != http.StatusNoContent {
		t.Fatalf("expected authenticated request to pass, got %d", authorized.Code)
	}
}

func TestValidatePolicyRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/policies/validate", strings.NewReader(`{"version":"v1","name":"test","unknown":true}`))
	rec := httptest.NewRecorder()
	(server{}).validatePolicy(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
