package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Prof-bilal/ShieldCI/backend/internal/policy"
)

// SQLiteCLI is a small local-development persistence adapter. It talks to the
// system sqlite3 executable so the API remains durable without binding the
// control plane to a CGO driver. The Store interface keeps this replaceable by
// a native SQLite or PostgreSQL adapter for production.
type SQLiteCLI struct{ path string }

func NewSQLite(path string) (*SQLiteCLI, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("sqlite path is required")
	}
	if _, err := exec.LookPath("sqlite3"); err != nil {
		return nil, fmt.Errorf("sqlite3 is unavailable: %w", err)
	}
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			return nil, fmt.Errorf("create sqlite directory: %w", err)
		}
	}
	s := &SQLiteCLI{path: path}
	if err := s.exec(context.Background(), `CREATE TABLE IF NOT EXISTS policies (id TEXT PRIMARY KEY, policy_json TEXT NOT NULL, created_at TEXT NOT NULL); CREATE TABLE IF NOT EXISTS jobs (id TEXT PRIMARY KEY, policy_id TEXT NOT NULL REFERENCES policies(id), command_json TEXT NOT NULL, status TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL); CREATE TABLE IF NOT EXISTS audit (id TEXT PRIMARY KEY, job_id TEXT NOT NULL REFERENCES jobs(id), type TEXT NOT NULL, decision TEXT NOT NULL, reason TEXT NOT NULL, created_at TEXT NOT NULL); CREATE TABLE IF NOT EXISTS findings (id TEXT PRIMARY KEY, job_id TEXT NOT NULL DEFAULT '', rule TEXT NOT NULL, severity TEXT NOT NULL, path TEXT NOT NULL, line INTEGER NOT NULL, message TEXT NOT NULL, fingerprint TEXT NOT NULL, created_at TEXT NOT NULL);`); err != nil {
		return nil, fmt.Errorf("initialize sqlite schema: %w", err)
	}
	return s, nil
}

func (s *SQLiteCLI) CreatePolicy(input policy.Policy) (PolicyRecord, error) {
	if err := input.Validate(); err != nil {
		return PolicyRecord{}, err
	}
	now := time.Now().UTC()
	record := PolicyRecord{ID: "pol_" + now.Format("20060102T150405.000000000"), Policy: input, CreatedAt: now}
	data, err := json.Marshal(input)
	if err != nil {
		return PolicyRecord{}, err
	}
	query := fmt.Sprintf("INSERT INTO policies(id, policy_json, created_at) VALUES(%s,%s,%s);", sqlQuote(record.ID), sqlQuote(string(data)), sqlQuote(now.Format(time.RFC3339Nano)))
	if err := s.exec(context.Background(), query); err != nil {
		return PolicyRecord{}, err
	}
	return record, nil
}

func (s *SQLiteCLI) GetPolicy(id string) (PolicyRecord, error) {
	var rows []struct {
		ID        string `json:"id"`
		Policy    string `json:"policy_json"`
		CreatedAt string `json:"created_at"`
	}
	if err := s.query(fmt.Sprintf("SELECT id, policy_json, created_at FROM policies WHERE id=%s;", sqlQuote(id)), &rows); err != nil {
		return PolicyRecord{}, err
	}
	if len(rows) == 0 {
		return PolicyRecord{}, ErrNotFound
	}
	var p policy.Policy
	if err := json.Unmarshal([]byte(rows[0].Policy), &p); err != nil {
		return PolicyRecord{}, fmt.Errorf("decode stored policy: %w", err)
	}
	created, err := time.Parse(time.RFC3339Nano, rows[0].CreatedAt)
	if err != nil {
		return PolicyRecord{}, fmt.Errorf("decode policy timestamp: %w", err)
	}
	return PolicyRecord{ID: rows[0].ID, Policy: p, CreatedAt: created}, nil
}

func (s *SQLiteCLI) CreateJob(policyID string, command []string) (Job, error) {
	if _, err := s.GetPolicy(policyID); err != nil {
		return Job{}, err
	}
	now := time.Now().UTC()
	job := Job{ID: "job_" + now.Format("20060102T150405.000000000"), PolicyID: policyID, Command: append([]string(nil), command...), Status: "queued", CreatedAt: now, UpdatedAt: now}
	data, err := json.Marshal(job.Command)
	if err != nil {
		return Job{}, err
	}
	query := fmt.Sprintf("INSERT INTO jobs(id, policy_id, command_json, status, created_at, updated_at) VALUES(%s,%s,%s,%s,%s,%s);", sqlQuote(job.ID), sqlQuote(policyID), sqlQuote(string(data)), sqlQuote(job.Status), sqlQuote(now.Format(time.RFC3339Nano)), sqlQuote(now.Format(time.RFC3339Nano)))
	if err := s.exec(context.Background(), query); err != nil {
		return Job{}, err
	}
	return job, nil
}

func (s *SQLiteCLI) GetJob(id string) (Job, error) {
	var rows []struct {
		ID        string `json:"id"`
		PolicyID  string `json:"policy_id"`
		Command   string `json:"command_json"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}
	if err := s.query(fmt.Sprintf("SELECT id, policy_id, command_json, status, created_at, updated_at FROM jobs WHERE id=%s;", sqlQuote(id)), &rows); err != nil {
		return Job{}, err
	}
	if len(rows) == 0 {
		return Job{}, ErrNotFound
	}
	var command []string
	if err := json.Unmarshal([]byte(rows[0].Command), &command); err != nil {
		return Job{}, fmt.Errorf("decode stored command: %w", err)
	}
	created, err := time.Parse(time.RFC3339Nano, rows[0].CreatedAt)
	if err != nil {
		return Job{}, err
	}
	updated, err := time.Parse(time.RFC3339Nano, rows[0].UpdatedAt)
	if err != nil {
		return Job{}, err
	}
	return Job{ID: rows[0].ID, PolicyID: rows[0].PolicyID, Command: command, Status: rows[0].Status, CreatedAt: created, UpdatedAt: updated}, nil
}

func (s *SQLiteCLI) ClaimNextQueued() (Job, error) {
	var rows []struct {
		ID string `json:"id"`
	}
	query := "UPDATE jobs SET status='running', updated_at=" + sqlQuote(time.Now().UTC().Format(time.RFC3339Nano)) + " WHERE id=(SELECT id FROM jobs WHERE status='queued' ORDER BY created_at LIMIT 1) RETURNING id;"
	if err := s.query(query, &rows); err != nil {
		return Job{}, err
	}
	if len(rows) == 0 {
		return Job{}, ErrNotFound
	}
	return s.GetJob(rows[0].ID)
}

func (s *SQLiteCLI) ListJobs() []Job {
	var rows []struct {
		ID        string `json:"id"`
		PolicyID  string `json:"policy_id"`
		Command   string `json:"command_json"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}
	if err := s.query("SELECT id, policy_id, command_json, status, created_at, updated_at FROM jobs ORDER BY created_at DESC;", &rows); err != nil {
		return []Job{}
	}
	jobs := make([]Job, 0, len(rows))
	for _, row := range rows {
		var command []string
		created, createdErr := time.Parse(time.RFC3339Nano, row.CreatedAt)
		updated, updatedErr := time.Parse(time.RFC3339Nano, row.UpdatedAt)
		if json.Unmarshal([]byte(row.Command), &command) == nil && createdErr == nil && updatedErr == nil {
			jobs = append(jobs, Job{ID: row.ID, PolicyID: row.PolicyID, Command: command, Status: row.Status, CreatedAt: created, UpdatedAt: updated})
		}
	}
	return jobs
}

func (s *SQLiteCLI) UpdateJobStatus(id, status string) (Job, error) {
	if _, err := s.GetJob(id); err != nil {
		return Job{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if err := s.exec(context.Background(), fmt.Sprintf("UPDATE jobs SET status=%s, updated_at=%s WHERE id=%s;", sqlQuote(status), sqlQuote(now), sqlQuote(id))); err != nil {
		return Job{}, err
	}
	return s.GetJob(id)
}

func (s *SQLiteCLI) AppendAudit(event AuditEvent) error {
	if _, err := s.GetJob(event.JobID); err != nil {
		return err
	}
	if event.ID == "" {
		event.ID = "evt_" + event.CreatedAt.Format("20060102T150405.000000000")
	}
	query := fmt.Sprintf("INSERT INTO audit(id, job_id, type, decision, reason, created_at) VALUES(%s,%s,%s,%s,%s,%s);", sqlQuote(event.ID), sqlQuote(event.JobID), sqlQuote(event.Type), sqlQuote(event.Decision), sqlQuote(event.Reason), sqlQuote(event.CreatedAt.UTC().Format(time.RFC3339Nano)))
	return s.exec(context.Background(), query)
}

func (s *SQLiteCLI) ListAudit(jobID string) ([]AuditEvent, error) {
	if _, err := s.GetJob(jobID); err != nil {
		return nil, err
	}
	var rows []AuditEvent
	if err := s.query(fmt.Sprintf("SELECT id, job_id, type, decision, reason, created_at FROM audit WHERE job_id=%s ORDER BY created_at ASC;", sqlQuote(jobID)), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *SQLiteCLI) CreateFinding(finding Finding) error {
	if finding.JobID != "" {
		if _, err := s.GetJob(finding.JobID); err != nil {
			return err
		}
	}
	if finding.ID == "" {
		finding.ID = "finding_" + finding.Fingerprint
	}
	if finding.CreatedAt.IsZero() {
		finding.CreatedAt = time.Now().UTC()
	}
	query := fmt.Sprintf("INSERT INTO findings(id, job_id, rule, severity, path, line, message, fingerprint, created_at) VALUES(%s,%s,%s,%s,%s,%d,%s,%s,%s);", sqlQuote(finding.ID), sqlQuote(finding.JobID), sqlQuote(finding.Rule), sqlQuote(finding.Severity), sqlQuote(finding.Path), finding.Line, sqlQuote(finding.Message), sqlQuote(finding.Fingerprint), sqlQuote(finding.CreatedAt.UTC().Format(time.RFC3339Nano)))
	return s.exec(context.Background(), query)
}

func (s *SQLiteCLI) ListFindings(jobID string) ([]Finding, error) {
	if jobID != "" {
		if _, err := s.GetJob(jobID); err != nil {
			return nil, err
		}
	}
	where := ""
	if jobID != "" {
		where = " WHERE job_id=" + sqlQuote(jobID)
	}
	var rows []Finding
	if err := s.query("SELECT id, job_id, rule, severity, path, line, message, fingerprint, created_at FROM findings"+where+" ORDER BY created_at DESC;", &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *SQLiteCLI) exec(ctx context.Context, query string) error {
	cmd := exec.CommandContext(ctx, "sqlite3", "-batch", s.path, query)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sqlite command failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func (s *SQLiteCLI) query(query string, target any) error {
	cmd := exec.Command("sqlite3", "-json", "-batch", s.path, query)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("sqlite query failed: %w", err)
	}
	if len(strings.TrimSpace(string(output))) == 0 {
		return nil
	}
	return json.Unmarshal(output, target)
}

func sqlQuote(value string) string { return "'" + strings.ReplaceAll(value, "'", "''") + "'" }
