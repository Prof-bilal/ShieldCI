package store

import (
	"errors"
	"sync"
	"time"

	"github.com/Prof-bilal/ShieldCI/backend/internal/policy"
)

var ErrNotFound = errors.New("record not found")

type PolicyRecord struct {
	ID        string        `json:"id"`
	Policy    policy.Policy `json:"policy"`
	CreatedAt time.Time     `json:"created_at"`
}

type Job struct {
	ID        string    `json:"id"`
	PolicyID  string    `json:"policy_id"`
	Command   []string  `json:"command"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuditEvent struct {
	ID        string    `json:"id"`
	JobID     string    `json:"job_id"`
	Type      string    `json:"type"`
	Decision  string    `json:"decision"`
	Reason    string    `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Finding struct {
	ID          string    `json:"id"`
	JobID       string    `json:"job_id,omitempty"`
	Rule        string    `json:"rule"`
	Severity    string    `json:"severity"`
	Path        string    `json:"path"`
	Line        int       `json:"line"`
	Message     string    `json:"message"`
	Fingerprint string    `json:"fingerprint"`
	CreatedAt   time.Time `json:"created_at"`
}

type Store interface {
	CreatePolicy(policy.Policy) (PolicyRecord, error)
	GetPolicy(string) (PolicyRecord, error)
	CreateJob(string, []string) (Job, error)
	ClaimNextQueued() (Job, error)
	GetJob(string) (Job, error)
	ListJobs() []Job
	UpdateJobStatus(string, string) (Job, error)
	AppendAudit(AuditEvent) error
	ListAudit(string) ([]AuditEvent, error)
	CreateFinding(Finding) error
	ListFindings(string) ([]Finding, error)
}

type Memory struct {
	mu       sync.RWMutex
	sequence uint64
	policies map[string]PolicyRecord
	jobs     map[string]Job
	audit    map[string][]AuditEvent
	findings []Finding
}

func NewMemory() *Memory {
	return &Memory{policies: map[string]PolicyRecord{}, jobs: map[string]Job{}, audit: map[string][]AuditEvent{}, findings: []Finding{}}
}

func (m *Memory) nextID(prefix string) string {
	m.sequence++
	return prefix + "_" + time.Now().UTC().Format("20060102T150405.000000000") + "_" + formatSequence(m.sequence)
}

func (m *Memory) CreatePolicy(input policy.Policy) (PolicyRecord, error) {
	if err := input.Validate(); err != nil {
		return PolicyRecord{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC()
	record := PolicyRecord{ID: m.nextID("pol"), Policy: input, CreatedAt: now}
	m.policies[record.ID] = record
	return record, nil
}

func (m *Memory) GetPolicy(id string) (PolicyRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	record, ok := m.policies[id]
	if !ok {
		return PolicyRecord{}, ErrNotFound
	}
	return record, nil
}

func (m *Memory) CreateJob(policyID string, command []string) (Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.policies[policyID]; !ok {
		return Job{}, ErrNotFound
	}
	now := time.Now().UTC()
	job := Job{ID: m.nextID("job"), PolicyID: policyID, Command: append([]string(nil), command...), Status: "queued", CreatedAt: now, UpdatedAt: now}
	m.jobs[job.ID] = job
	return job, nil
}

func (m *Memory) GetJob(id string) (Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	job, ok := m.jobs[id]
	if !ok {
		return Job{}, ErrNotFound
	}
	return job, nil
}

func (m *Memory) ClaimNextQueued() (Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var selected *Job
	for _, candidate := range m.jobs {
		if candidate.Status == "queued" && (selected == nil || candidate.CreatedAt.Before(selected.CreatedAt)) {
			copy := candidate
			selected = &copy
		}
	}
	if selected == nil {
		return Job{}, ErrNotFound
	}
	selected.Status = "running"
	selected.UpdatedAt = time.Now().UTC()
	m.jobs[selected.ID] = *selected
	return *selected, nil
}

func (m *Memory) ListJobs() []Job {
	m.mu.RLock()
	defer m.mu.RUnlock()
	jobs := make([]Job, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

func (m *Memory) UpdateJobStatus(id, status string) (Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[id]
	if !ok {
		return Job{}, ErrNotFound
	}
	job.Status = status
	job.UpdatedAt = time.Now().UTC()
	m.jobs[id] = job
	return job, nil
}

func (m *Memory) AppendAudit(event AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.jobs[event.JobID]; !ok {
		return ErrNotFound
	}
	if event.ID == "" {
		event.ID = m.nextID("evt")
	}
	m.audit[event.JobID] = append(m.audit[event.JobID], event)
	return nil
}

func (m *Memory) ListAudit(jobID string) ([]AuditEvent, error) {
	if _, err := m.GetJob(jobID); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]AuditEvent(nil), m.audit[jobID]...), nil
}

func (m *Memory) CreateFinding(finding Finding) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if finding.JobID != "" {
		if _, ok := m.jobs[finding.JobID]; !ok {
			return ErrNotFound
		}
	}
	if finding.ID == "" {
		finding.ID = m.nextID("finding")
	}
	if finding.CreatedAt.IsZero() {
		finding.CreatedAt = time.Now().UTC()
	}
	m.findings = append(m.findings, finding)
	return nil
}

func (m *Memory) ListFindings(jobID string) ([]Finding, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if jobID != "" {
		if _, ok := m.jobs[jobID]; !ok {
			return nil, ErrNotFound
		}
	}
	findings := make([]Finding, 0, len(m.findings))
	for _, finding := range m.findings {
		if jobID == "" || finding.JobID == jobID {
			findings = append(findings, finding)
		}
	}
	return findings, nil
}

func formatSequence(n uint64) string {
	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	if n == 0 {
		return "0"
	}
	var out [13]byte
	i := len(out)
	for n > 0 {
		i--
		out[i] = digits[n%36]
		n /= 36
	}
	return string(out[i:])
}
