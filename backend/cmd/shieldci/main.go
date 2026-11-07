package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Prof-bilal/ShieldCI/backend/internal/policy"
	"github.com/Prof-bilal/ShieldCI/backend/internal/runner"
	"github.com/Prof-bilal/ShieldCI/backend/internal/scanner"
	"github.com/Prof-bilal/ShieldCI/backend/internal/store"
	"github.com/Prof-bilal/ShieldCI/backend/internal/wardenadapter"
)

type server struct{ store store.Store }

func (server) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "shieldci-control-plane"})
}

func (server) validatePolicy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}
	defer r.Body.Close()
	var input policy.Policy
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body is not valid policy JSON")
		return
	}
	if err := input.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_policy", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"valid": true, "policy": input})
}

func (s server) createPolicy(w http.ResponseWriter, r *http.Request) {
	input, ok := decodePolicy(w, r)
	if !ok {
		return
	}
	record, err := s.store.CreatePolicy(input)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_policy", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, record)
}

func (s server) policies(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusInternalServerError, "store_unavailable", "store is not configured")
		return
	}
	if r.Method == http.MethodPost && r.URL.Path == "/api/v1/policies" {
		s.createPolicy(w, r)
		return
	}
	if r.Method == http.MethodGet && r.URL.Path == "/api/v1/policies" {
		writeJSON(w, http.StatusOK, map[string]any{"policies": s.store.ListPolicies()})
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/policies/")
	if r.Method != http.MethodGet || id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET for a policy id")
		return
	}
	record, err := s.store.GetPolicy(id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "policy not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "store_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, record)
}

type createJobRequest struct {
	PolicyID string   `json:"policy_id"`
	Command  []string `json:"command"`
}

func (s server) jobs(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusInternalServerError, "store_unavailable", "store is not configured")
		return
	}
	if r.Method == http.MethodPost && r.URL.Path == "/api/v1/jobs" {
		defer r.Body.Close()
		var input createJobRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || input.PolicyID == "" || len(input.Command) == 0 {
			writeError(w, http.StatusBadRequest, "invalid_job", "policy_id and a non-empty command are required")
			return
		}
		job, err := s.store.CreateJob(input.PolicyID, input.Command)
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "policy not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
		writeJSON(w, http.StatusAccepted, job)
		return
	}
	if r.Method == http.MethodGet && r.URL.Path == "/api/v1/jobs" {
		writeJSON(w, http.StatusOK, map[string]any{"jobs": s.store.ListJobs()})
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/jobs/")
	parts := strings.Split(strings.Trim(id, "/"), "/")
	if len(parts) == 2 && parts[1] == "audit" && r.Method == http.MethodGet {
		events, err := s.store.ListAudit(parts[0])
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "job not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"events": events})
		return
	}
	if len(parts) == 2 && parts[1] == "findings" && r.Method == http.MethodGet {
		findings, err := s.store.ListFindings(parts[0])
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "job not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"findings": findings})
		return
	}
	if r.Method != http.MethodGet || len(parts) != 1 || parts[0] == "" {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "unsupported jobs route")
		return
	}
	job, err := s.store.GetJob(parts[0])
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "job not found")
		return
	}
	writeJSON(w, http.StatusOK, job)
}

type scanSecretsRequest struct {
	Root  string `json:"root"`
	JobID string `json:"job_id"`
}

func (s server) findings(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusInternalServerError, "store_unavailable", "store is not configured")
		return
	}
	if r.Method == http.MethodGet && r.URL.Path == "/api/v1/findings" {
		findings, err := s.store.ListFindings(r.URL.Query().Get("job_id"))
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "job not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"findings": findings})
		return
	}
	if r.Method != http.MethodPost || r.URL.Path != "/api/v1/scans/secrets" {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST /api/v1/scans/secrets or GET /api/v1/findings")
		return
	}
	defer r.Body.Close()
	var input scanSecretsRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body is not valid scan input")
		return
	}
	if input.Root == "" {
		input.Root = "."
	}
	if input.JobID != "" {
		if _, err := s.store.GetJob(input.JobID); errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "job not found")
			return
		}
	}
	results, err := scanner.ScanSecrets(r.Context(), input.Root)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "scan_failed", err.Error())
		return
	}
	for _, result := range results {
		if err := s.store.CreateFinding(store.Finding{ID: result.ID, JobID: input.JobID, Rule: result.Rule, Severity: result.Severity, Path: result.Path, Line: result.Line, Message: result.Message, Fingerprint: result.Fingerprint}); err != nil {
			writeError(w, http.StatusInternalServerError, "store_error", err.Error())
			return
		}
	}
	writeJSON(w, http.StatusCreated, map[string]any{"findings": results, "count": len(results)})
}

func decodePolicy(w http.ResponseWriter, r *http.Request) (policy.Policy, bool) {
	defer r.Body.Close()
	var input policy.Policy
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body is not valid policy JSON")
		return policy.Policy{}, false
	}
	if err := input.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_policy", err.Error())
		return policy.Policy{}, false
	}
	return input, true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func authMiddleware(next http.Handler, expectedToken string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if expectedToken == "" || r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		value := strings.TrimSpace(r.Header.Get("Authorization"))
		if value != "Bearer "+expectedToken {
			writeError(w, http.StatusUnauthorized, "unauthorized", "a valid bearer token is required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": code, "message": message})
}

func main() {
	addr := os.Getenv("SHIELDCI_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	databasePath := os.Getenv("SHIELDCI_DB")
	if databasePath == "" {
		databasePath = ".shieldci/shieldci.db"
	}
	database, err := store.NewSQLite(databasePath)
	if err != nil {
		log.Fatalf("initialize ShieldCI store: %v", err)
	}
	svc := server{store: database}
	if binary := os.Getenv("SHIELDCI_WARDEN_BIN"); binary != "" {
		adapter, err := wardenadapter.New(wardenadapter.Config{BinaryPath: binary, Backend: os.Getenv("SHIELDCI_WARDEN_BACKEND"), WorkingDir: "."})
		if err != nil {
			log.Fatalf("initialize Warden adapter: %v", err)
		}
		go runWorker(context.Background(), runner.Runner{Store: database, Adapter: adapter})
	}
	mux := newMux(svc)
	log.Printf("ShieldCI control plane listening on %s", addr)
	if err := http.ListenAndServe(addr, authMiddleware(mux, os.Getenv("SHIELDCI_API_TOKEN"))); err != nil {
		log.Fatal(err)
	}
}

func newMux(svc server) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", svc.health)
	mux.HandleFunc("/api/v1/policies/validate", svc.validatePolicy)
	mux.HandleFunc("/api/v1/policies", svc.policies)
	mux.HandleFunc("/api/v1/policies/", svc.policies)
	mux.HandleFunc("/api/v1/jobs", svc.jobs)
	mux.HandleFunc("/api/v1/jobs/", svc.jobs)
	mux.HandleFunc("/api/v1/findings", svc.findings)
	mux.HandleFunc("/api/v1/scans/secrets", svc.findings)
	return mux
}

func runWorker(ctx context.Context, jobRunner runner.Runner) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		_, err := jobRunner.RunNext(ctx)
		if err != nil {
			log.Printf("ShieldCI worker: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
