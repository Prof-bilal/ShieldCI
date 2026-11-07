# Warden CI/CD - MVP Technical Architecture

## 📋 Table of Contents
1. [Architecture Overview](#overview)
2. [High-Level Architecture](#high-level)
3. [Component Architecture](#components)
4. [Data Models](#data-models)
5. [API Design](#api)
6. [Integration Points](#integrations)
7. [Security Model](#security)
8. [Deployment Architecture](#deployment)

---

## <a name="overview"></a>1. Architecture Overview

### 1.1 Design Principles

**Security Principles:**
- ✅ **Deny by Default** - All access denied unless explicitly allowed
- ✅ **Least Privilege** - Minimum required access for each process
- ✅ **Zero Trust** - Never trust, always verify
- ✅ **Defense in Depth** - Multiple layers of security
- ✅ **Fail Secure** - Deny on failure, not allow

**Design Principles:**
- ✅ **Simplicity** - Keep it simple and focused
- ✅ **Developer-Friendly** - Easy to use and understand
- ✅ **Extensibility** - Built for future growth
- ✅ **Performance** - Minimal overhead
- ✅ **Reliability** - High availability and fault tolerance

### 1.2 Technology Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| **Backend** | Go 1.21+ | Core application logic, API server |
| **Frontend** | React + TypeScript | Web dashboard |
| **Database** | PostgreSQL 15+ | Structured data storage |
| **Cache/Queue** | Redis | Job queue, caching |
| **Storage** | S3/Cloud Storage | Logs, artifacts |
| **Sandboxing** | Warden Core | OS-level isolation |
| **Scanning** | Semgrep | SAST scanning |
| **CI/CD** | GitHub Actions | CI/CD integration |
| **Cloud** | AWS/GCP | Hosting infrastructure |
| **Monitoring** | Prometheus + Grafana | Metrics and monitoring |
| **Logging** | ELK/CloudWatch | Log management |

---

## <a name="high-level"></a>2. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        WARDEN CI/CD MVP                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────────────┐  │
│  │   User     │    │  GitHub    │    │      CONTROL PLANE          │  │
│  │  Interface │    │  Actions   │    │                             │  │
│  │             │    │             │    │  ┌───────────────────────┐  │  │
│  └──────┬──────┘    └──────┬──────┘    │  │     API Server        │  │  │
│         │                │            │  └──────────┬────────────┘  │  │
│         │                │            │             │              │  │
│         │                │            │  ┌──────────┴────────────┐  │  │
│         │                │            │  │   Policy Engine       │  │  │
│         │                │            │  └──────────┬────────────┘  │  │
│         │                │            │             │              │  │
│         │                │            │  ┌──────────▼────────────┐  │  │
│         │                │            │  │   Job Scheduler       │  │  │
│         │                │            │  └──────────┬────────────┘  │  │
│         │                │            │             │              │  │
│         │                │            │  ┌──────────▼────────────┐  │  │
│         │                │            │  │   Audit Engine        │  │  │
│         │                │            └─────────────────────────────┘  │
│         │                │                                          │
│         │                ▼            ┌─────────────────────────────┐  │
│         │            ┌───▼─────┐    │      DATA PLANE              │  │
│         │            │ GitHub  │    │                             │  │
│         │            │ Action  │    │  ┌───────────────────────┐  │  │
│         │            └───┬─────┘    │  │   Sandbox Worker     │  │  │
│         │                │         │  │   (Warden Core)      │  │  │
│         │                │         │  └──────────┬────────────┘  │  │
│         │                │         │             │              │  │
│         │                │         │  ┌──────────▼────────────┐  │  │
│         │                │         │  │   Scanning Worker    │  │  │
│         │                │         │  │   (Semgrep, Secrets)  │  │  │
│         │                │         │  └───────────────────────┘  │  │
│         │                │         └─────────────────────────────┘  │
│         │                │                                          │
│         ▼                ▼            ┌─────────────────────────────┐  │
│  ┌─────────────┐  ┌─────────────┐    │      STORAGE                │  │
│  │   Web      │  │   CLI      │    │                             │  │
│  │  Dashboard │  │   Tool     │    │  ┌───────────────────────┐  │  │
│  └─────────────┘  └─────────────┘    │  │  PostgreSQL         │  │  │
│                                          │  │  (Policies, Jobs)   │  │  │
│                                          │  └───────────────────────┘  │  │
│                                          │                             │  │
│                                          │  ┌───────────────────────┐  │  │
│                                          │  │  S3 / Object Storage │  │  │
│                                          │  │  (Logs, Artifacts)   │  │  │
│                                          │  └───────────────────────┘  │  │
│                                          └─────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

---

## <a name="components"></a>3. Component Architecture

### 3.1 Control Plane Components

#### 1. API Server
**Purpose**: REST API for all external interactions

**Responsibilities**:
- Handle all HTTP requests
- Authentication and authorization
- Request validation
- Business logic
- Error handling

**Technical Details**:
- **Language**: Go (Gin or Echo framework)
- **Authentication**: JWT, OAuth 2.0
- **API Style**: RESTful
- **Documentation**: OpenAPI/Swagger

**Endpoints**:
```
/api/v1
├── /auth
│   ├── POST /login          - User login
│   ├── POST /logout         - User logout
│   ├── POST /refresh        - Refresh token
│   └── GET /me              - Get current user
├── /organizations
│   ├── GET /                - List organizations
│   ├── POST /               - Create organization
│   ├── GET /:id             - Get organization
│   ├── PUT /:id             - Update organization
│   └── DELETE /:id          - Delete organization
├── /policies
│   ├── GET /                - List policies
│   ├── POST /               - Create policy
│   ├── GET /:id             - Get policy
│   ├── PUT /:id             - Update policy
│   ├── DELETE /:id          - Delete policy
│   └── POST /:id/test       - Test policy
├── /jobs
│   ├── GET /                - List jobs
│   ├── POST /               - Create job
│   ├── GET /:id             - Get job
│   ├── GET /:id/logs        - Get job logs
│   └── GET /:id/artifacts   - Get job artifacts
├── /scan
│   ├── POST /sast           - Run SAST scan
│   ├── POST /secrets        - Run secrets scan
│   └── GET /results         - Get scan results
├── /audit
│   ├── GET /logs            - Get audit logs
│   └── GET /logs/export     - Export audit logs
└── /alerts
    ├── GET /                - List alerts
    ├── POST /               - Create alert
    ├── GET /:id             - Get alert
    └── PUT /:id             - Update alert
```

**Implementation**:
```go
package main

import (
    "github.com/gin-gonic/gin"
    "warden-cicd/internal/handlers"
    "warden-cicd/internal/middleware"
)

func main() {
    r := gin.Default()
    
    // Middleware
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())
    r.Use(middleware.CORS())
    
    // Authentication
    auth := r.Group("/api/v1/auth")
    auth.POST("/login", handlers.Login)
    auth.POST("/refresh", handlers.Refresh)
    
    // Protected routes
    api := r.Group("/api/v1", middleware.Auth())
    {
        api.GET("/me", handlers.GetMe)
        
        // Organizations
        orgs := api.Group("/organizations")
        orgs.GET("/", handlers.ListOrgs)
        orgs.POST("/", handlers.CreateOrg)
        
        // Policies
        policies := api.Group("/policies")
        policies.GET("/", handlers.ListPolicies)
        policies.POST("/", handlers.CreatePolicy)
        
        // Jobs
        jobs := api.Group("/jobs")
        jobs.GET("/", handlers.ListJobs)
        jobs.POST("/", handlers.CreateJob)
    }
    
    r.Run(":8080")
}
```

---

#### 2. Policy Engine
**Purpose**: Parse, validate, and manage security policies

**Responsibilities**:
- Parse YAML policy files
- Validate policy syntax and semantics
- Store and retrieve policies
- Version policies
- Resolve policy references
- Generate compiled policies for sandboxes

**Technical Details**:
- **Language**: Go
- **Storage**: PostgreSQL
- **Format**: YAML

**Policy Schema**:
```go
type Policy struct {
    Version     string            `yaml:"version"`
    Name        string            `yaml:"name"`
    Description string            `yaml:"description"`
    Filesystem  FilesystemPolicy  `yaml:"filesystem"`
    Network     NetworkPolicy     `yaml:"network"`
    Env         EnvPolicy         `yaml:"env"`
    Process     ProcessPolicy     `yaml:"process"`
    Limits      LimitPolicy       `yaml:"limits"`
}

type FilesystemPolicy struct {
    Read  []string `yaml:"read"`
    Write []string `yaml:"write"`
    Deny  []string `yaml:"deny"`
}

type NetworkPolicy struct {
    Allow []string `yaml:"allow"`
    Deny  []string `yaml:"deny"`
}

type EnvPolicy struct {
    Allow []string `yaml:"allow"`
    Deny  []string `yaml:"deny"`
}

type ProcessPolicy struct {
    Allow []string `yaml:"allow"`
    Deny  []string `yaml:"deny"`
}

type LimitPolicy struct {
    MemoryMB int `yaml:"memory_mb"`
    TimeoutS int `yaml:"timeout_s"`
    CPUCores int `yaml:"cpu_cores"`
}
```

**Implementation**:
```go
package policy

import (
    "gopkg.in/yaml.v3"
    "warden-cicd/internal/models"
)

func Parse(policyYAML []byte) (*models.Policy, error) {
    var policy models.Policy
    if err := yaml.Unmarshal(policyYAML, &policy); err != nil {
        return nil, err
    }
    
    // Validate
    if err := validate(&policy); err != nil {
        return nil, err
    }
    
    return &policy, nil
}

func validate(policy *models.Policy) error {
    // Validate version
    if policy.Version != "1.0" {
        return fmt.Errorf("unsupported policy version: %s", policy.Version)
    }
    
    // Validate filesystem
    if err := validateFilesystem(&policy.Filesystem); err != nil {
        return err
    }
    
    // Validate network
    if err := validateNetwork(&policy.Network); err != nil {
        return err
    }
    
    // ... more validation
    
    return nil
}
```

---

#### 3. Job Scheduler
**Purpose**: Schedule and manage CI/CD jobs

**Responsibilities**:
- Receive job requests from API
- Queue jobs for execution
- Allocate jobs to sandbox workers
- Track job status
- Handle job retries
- Store job results

**Technical Details**:
- **Language**: Go
- **Queue**: Redis (for job queue)
- **Storage**: PostgreSQL (for job metadata)

**Job Model**:
```go
type Job struct {
    ID          string       `json:"id"`
    UserID      string       `json:"user_id"`
    OrgID       string       `json:"org_id"`
    PolicyID    string       `json:"policy_id"`
    Workflow    string       `json:"workflow"`  // GitHub workflow file
    Command     string       `json:"command"`   // Command to run
    Status      JobStatus    `json:"status"`    // pending, running, completed, failed
    Result      JobResult    `json:"result"`    // exit code, output, etc.
    CreatedAt   time.Time    `json:"created_at"`
    StartedAt   *time.Time   `json:"started_at"`
    CompletedAt *time.Time   `json:"completed_at"`
    Duration    time.Duration `json:"duration"`
}

type JobStatus string

const (
    JobStatusPending  JobStatus = "pending"
    JobStatusRunning  JobStatus = "running"
    JobStatusCompleted JobStatus = "completed"
    JobStatusFailed   JobStatus = "failed"
)

type JobResult struct {
    ExitCode    int    `json:"exit_code"`
    Stdout      string `json:"stdout"`
    Stderr      string `json:"stderr"`
    Logs        string `json:"logs"`
    Artifacts   string `json:"artifacts"` // URL to artifacts
}
```

**Implementation**:
```go
package scheduler

import (
    "context"
    "github.com/go-redis/redis/v8"
    "warden-cicd/internal/models"
)

type Scheduler struct {
    redisClient *redis.Client
    jobRepo     JobRepository
    workerPool  WorkerPool
}

func (s *Scheduler) Enqueue(ctx context.Context, job *models.Job) error {
    // Store job in database
    if err := s.jobRepo.Create(job); err != nil {
        return err
    }
    
    // Add to queue
    return s.redisClient.LPush(ctx, "jobs", job.ID).Err()
}

func (s *Scheduler) Start(ctx context.Context) error {
    // Subscribe to queue
    pubsub := s.redisClient.Subscribe(ctx, "jobs")
    ch := pubsub.Channel()
    
    for msg := range ch {
        jobID := msg.Payload
        job, err := s.jobRepo.Get(jobID)
        if err != nil {
            continue
        }
        
        // Find available worker
        worker := s.workerPool.GetAvailable()
        if worker == nil {
            // Requeue
            s.redisClient.RPush(ctx, "jobs", jobID)
            continue
        }
        
        // Assign job to worker
        if err := worker.AssignJob(job); err != nil {
            // Mark as failed
            job.Status = models.JobStatusFailed
            s.jobRepo.Update(job)
        }
    }
    
    return nil
}
```

---

#### 4. Audit Engine
**Purpose**: Collect, store, and query audit logs

**Responsibilities**:
- Receive audit events from all components
- Store audit logs in database and object storage
- Provide search and filtering capabilities
- Export audit logs
- Archive old logs

**Technical Details**:
- **Storage**: PostgreSQL (for structured metadata) + S3 (for raw logs)
- **Format**: JSON (structured)
- **Retention**: 30 days hot storage, 90 days warm storage

**Audit Event Model**:
```go
type AuditEvent struct {
    ID            string            `json:"id"`
    Timestamp     time.Time         `json:"timestamp"`
    SessionID     string            `json:"session_id"`
    JobID         string            `json:"job_id"`
    UserID        string            `json:"user_id"`
    OrgID         string            `json:"org_id"`
    EventType     AuditEventType    `json:"event_type"`
    Action        string            `json:"action"`
    Resource      string            `json:"resource"`
    Allowed       bool              `json:"allowed"`
    PolicyID      string            `json:"policy_id"`
    PolicyRule    string            `json:"policy_rule"`
    Details       map[string]string `json:"details"`
    Metadata      map[string]string `json:"metadata"`
}

type AuditEventType string

const (
    EventTypeFilesystem AuditEventType = "filesystem"
    EventTypeNetwork   AuditEventType = "network"
    EventTypeProcess   AuditEventType = "process"
    EventTypeEnv       AuditEventType = "env"
    EventTypeJob       AuditEventType = "job"
    EventTypePolicy    AuditEventType = "policy"
    EventTypeSecurity  AuditEventType = "security"
)
```

---

### 3.2 Data Plane Components

#### 1. Sandbox Worker
**Purpose**: Execute jobs in Warden sandboxes

**Responsibilities**:
- Receive jobs from job scheduler
- Create Warden sandboxes
- Apply policies to sandboxes
- Execute commands in sandboxes
- Monitor sandbox execution
- Collect results and logs
- Clean up sandboxes

**Technical Details**:
- **Language**: Go
- **Integration**: Warden core library (Go API)
- **Platforms**: Linux, macOS, Windows

**Implementation**:
```go
package worker

import (
    "context"
    "warden-cicd/internal/models"
    "warden-starter/warden" // Warden core library
)

type SandboxWorker struct {
    id       string
    policy   *models.Policy
    sandbox  warden.Sandbox
    jobQueue chan *models.Job
    resultChan chan JobResult
}

func (w *SandboxWorker) CreateSandbox(ctx context.Context, policy *models.Policy) error {
    // Convert policy to Warden config
    config := convertPolicyToWardenConfig(policy)
    
    // Create sandbox
    sb, err := warden.NewSandbox(config)
    if err != nil {
        return err
    }
    
    w.sandbox = sb
    w.policy = policy
    return nil
}

func (w *SandboxWorker) Execute(ctx context.Context, job *models.Job) (*models.JobResult, error) {
    // Create sandbox
    if err := w.CreateSandbox(ctx, job.Policy); err != nil {
        return nil, err
    }
    defer w.sandbox.Destroy()
    
    // Execute command in sandbox
    cmd := w.sandbox.Exec(job.Command)
    output, err := cmd.CombinedOutput()
    
    result := &models.JobResult{
        ExitCode: cmd.ProcessState.ExitCode(),
        Stdout:   string(output),
    }
    
    if err != nil {
        result.Stderr = err.Error()
    }
    
    return result, nil
}

func convertPolicyToWardenConfig(policy *models.Policy) *warden.Config {
    // Convert Warden CI/CD policy to Warden core config
    config := &warden.Config{
        Filesystem: &warden.FilesystemConfig{
            Read:  policy.Filesystem.Read,
            Write: policy.Filesystem.Write,
            Deny:  policy.Filesystem.Deny,
        },
        Network: &warden.NetworkConfig{
            Allow: policy.Network.Allow,
            Deny:  policy.Network.Deny,
        },
        Env: &warden.EnvConfig{
            Allow: policy.Env.Allow,
            Deny:  policy.Env.Deny,
        },
        Limits: &warden.LimitConfig{
            MemoryMB: policy.Limits.MemoryMB,
            TimeoutS: policy.Limits.TimeoutS,
        },
    }
    
    return config
}
```

---

#### 2. Scanning Worker
**Purpose**: Perform security scanning of code

**Responsibilities**:
- Receive scan requests
- Run SAST scans (Semgrep)
- Run secrets detection
- Collect scan results
- Store scan history
- Report findings

**Technical Details**:
- **Language**: Go
- **Tools**: Semgrep (SAST), custom patterns (secrets)
- **Execution**: Run scans in Warden sandboxes

**Implementation**:
```go
package scanner

import (
    "context"
    "os/exec"
    "warden-cicd/internal/models"
)

type Scanner struct {
    semgrepPath string
    patternsPath string
}

func (s *Scanner) RunSAST(ctx context.Context, dir string) (*models.ScanResult, error) {
    // Run Semgrep in sandbox
    cmd := exec.Command("semgrep", "scan", dir, "--config", "auto", "--json")
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    // Parse results
    result, err := parseSemgrepOutput(output)
    if err != nil {
        return nil, err
    }
    
    return result, nil
}

func (s *Scanner) RunSecretsScan(ctx context.Context, dir string) (*models.ScanResult, error) {
    // Run custom secrets detection
    cmd := exec.Command("warden-cicd", "secrets", "scan", dir)
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    // Parse results
    result, err := parseSecretsOutput(output)
    if err != nil {
        return nil, err
    }
    
    return result, nil
}
```

---

#### 3. CLI Tool
**Purpose**: Command-line interface for Warden CI/CD

**Technical Details**:
- **Language**: Go
- **Framework**: Cobra
- **Commands**: run, scan, policy, auth, etc.

**Implementation**:
```go
package main

import (
    "github.com/spf13/cobra"
    "warden-cicd/internal/commands"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "warden-cicd",
        Short: "Warden CI/CD - Zero-Trust CI/CD Security",
    }
    
    // Add commands
    rootCmd.AddCommand(
        commands.RunCmd,
        commands.ScanCmd,
        commands.PolicyCmd,
        commands.AuthCmd,
    )
    
    if err := rootCmd.Execute(); err != nil {
        panic(err)
    }
}

// commands/run.go
package commands

import (
    "github.com/spf13/cobra"
    "warden-cicd/internal/client"
)

var RunCmd = &cobra.Command{
    Use:   "run [command]",
    Short: "Run a command in Warden sandbox",
    RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) < 1 {
            return cmd.Usage()
        }
        
        policy, _ := cmd.Flags().GetString("policy")
        
        c := client.NewClient()
        job := &models.Job{
            Command:  args[0],
            PolicyID: policy,
        }
        
        result, err := c.RunJob(job)
        if err != nil {
            return err
        }
        
        cmd.Println(result.Stdout)
        return nil
    },
}

func init() {
    RunCmd.Flags().StringP("policy", "p", "", "Policy file or ID")
}
```

---

#### 4. Web Dashboard
**Purpose**: Web interface for monitoring and management

**Technical Details**:
- **Frontend**: React + TypeScript
- **Backend**: Go REST API
- **Styling**: Tailwind CSS
- **Authentication**: OAuth 2.0

**Page Structure**:
```
src/
├── components/
│   ├── layout/
│   │   ├── Header.tsx
│   │   ├── Sidebar.tsx
│   │   └── Footer.tsx
│   ├── policies/
│   │   ├── PolicyList.tsx
│   │   ├── PolicyEditor.tsx
│   │   └── PolicyCard.tsx
│   ├── jobs/
│   │   ├── JobList.tsx
│   │   ├── JobDetail.tsx
│   │   └── JobLogs.tsx
│   └── security/
│       ├── ScanResults.tsx
│       └── AuditLogs.tsx
├── pages/
│   ├── Dashboard.tsx
│   ├── Policies.tsx
│   ├── Jobs.tsx
│   ├── Security.tsx
│   ├── Audit.tsx
│   ├── Alerts.tsx
│   └── Settings.tsx
├── hooks/
│   ├── useAuth.ts
│   ├── useJobs.ts
│   └── usePolicies.ts
└── App.tsx
```

**Example Component**:
```tsx
// components/policies/PolicyEditor.tsx
import React, { useState } from 'react';
import { Policy } from '../../types';

const PolicyEditor: React.FC<{ policy: Policy }> = ({ policy }) => {
  const [yaml, setYaml] = useState<string>(policyYaml);
  const [errors, setErrors] = useState<string[]>([]);

  const handleValidate = async () => {
    try {
      await validatePolicy(yaml);
      setErrors([]);
    } catch (err) {
      setErrors(err.message.split('\n'));
    }
  };

  const handleSave = async () => {
    await savePolicy(policy.id, yaml);
  };

  return (
    <div className="policy-editor">
      <div className="editor-header">
        <h2>Policy Editor</h2>
        <div className="editor-actions">
          <button onClick={handleValidate}>Validate</button>
          <button onClick={handleSave} disabled={errors.length > 0}>
            Save
          </button>
        </div>
      </div>
      
      <div className="editor-container">
        <textarea
          value={yaml}
          onChange={(e) => setYaml(e.target.value)}
          className="yaml-editor"
        />
        
        {errors.length > 0 && (
          <div className="errors">
            <h3>Validation Errors:</h3>
            <ul>
              {errors.map((error, index) => (
                <li key={index}>{error}</li>
              ))}
            </ul>
          </div>
        )}
      </div>
      
      <div className="editor-preview">
        <h3>Policy Summary</h3>
        <PolicySummary yaml={yaml} />
      </div>
    </div>
  );
};

export default PolicyEditor;
```

---

## <a name="data-models"></a>4. Data Models

### 4.1 Database Schema

#### PostgreSQL Tables

```sql
-- Users
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Organizations
CREATE TABLE organizations (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Organization Members
CREATE TABLE organization_members (
    id VARCHAR(36) PRIMARY KEY,
    org_id VARCHAR(36) REFERENCES organizations(id) ON DELETE CASCADE,
    user_id VARCHAR(36) REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL, -- owner, admin, member
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Policies
CREATE TABLE policies (
    id VARCHAR(36) PRIMARY KEY,
    org_id VARCHAR(36) REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    yaml TEXT NOT NULL,
    version VARCHAR(50) NOT NULL DEFAULT '1.0',
    is_template BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Policy Versions
CREATE TABLE policy_versions (
    id VARCHAR(36) PRIMARY KEY,
    policy_id VARCHAR(36) REFERENCES policies(id) ON DELETE CASCADE,
    yaml TEXT NOT NULL,
    version INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Jobs
CREATE TABLE jobs (
    id VARCHAR(36) PRIMARY KEY,
    org_id VARCHAR(36) REFERENCES organizations(id) ON DELETE CASCADE,
    user_id VARCHAR(36) REFERENCES users(id) ON DELETE SET NULL,
    policy_id VARCHAR(36) REFERENCES policies(id) ON DELETE SET NULL,
    workflow_file TEXT,
    command TEXT,
    status VARCHAR(50) NOT NULL, -- pending, running, completed, failed
    exit_code INTEGER,
    stdout TEXT,
    stderr TEXT,
    logs_url VARCHAR(255),
    artifacts_url VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    duration INTERVAL
);

-- Job Results
CREATE TABLE job_results (
    id VARCHAR(36) PRIMARY KEY,
    job_id VARCHAR(36) REFERENCES jobs(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Scan Results
CREATE TABLE scan_results (
    id VARCHAR(36) PRIMARY KEY,
    job_id VARCHAR(36) REFERENCES jobs(id) ON DELETE CASCADE,
    scan_type VARCHAR(50) NOT NULL, -- sast, secrets
    tool VARCHAR(50) NOT NULL,
    findings JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Audit Logs
CREATE TABLE audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    org_id VARCHAR(36) REFERENCES organizations(id) ON DELETE CASCADE,
    user_id VARCHAR(36) REFERENCES users(id) ON DELETE SET NULL,
    job_id VARCHAR(36) REFERENCES jobs(id) ON DELETE SET NULL,
    session_id VARCHAR(36),
    event_type VARCHAR(50) NOT NULL, -- filesystem, network, process, env, job, policy
    action VARCHAR(50) NOT NULL,
    resource VARCHAR(255) NOT NULL,
    allowed BOOLEAN NOT NULL,
    policy_id VARCHAR(36) REFERENCES policies(id) ON DELETE SET NULL,
    policy_rule VARCHAR(255),
    details JSONB,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Alerts
CREATE TABLE alerts (
    id VARCHAR(36) PRIMARY KEY,
    org_id VARCHAR(36) REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    alert_type VARCHAR(50) NOT NULL, -- policy_violation, security_finding, job_failure
    severity VARCHAR(50) NOT NULL, -- low, medium, high, critical
    is_enabled BOOLEAN DEFAULT TRUE,
    config JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Alert Deliveries
CREATE TABLE alert_deliveries (
    id VARCHAR(36) PRIMARY KEY,
    alert_id VARCHAR(36) REFERENCES alerts(id) ON DELETE CASCADE,
    org_id VARCHAR(36) REFERENCES organizations(id) ON DELETE CASCADE,
    channel VARCHAR(50) NOT NULL, -- email, webhook, in_app
    recipient VARCHAR(255),
    status VARCHAR(50) NOT NULL, -- pending, sent, failed
    data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- GitHub Integration
CREATE TABLE github_integrations (
    id VARCHAR(36) PRIMARY KEY,
    org_id VARCHAR(36) REFERENCES organizations(id) ON DELETE CASCADE,
    installation_id VARCHAR(255) UNIQUE NOT NULL,
    access_token TEXT NOT NULL,
    token_expires_at TIMESTAMP WITH TIME ZONE,
    refresh_token TEXT,
    webhook_secret VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_jobs_org_id ON jobs(org_id);
CREATE INDEX idx_jobs_user_id ON jobs(user_id);
CREATE INDEX idx_jobs_policy_id ON jobs(policy_id);
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_created_at ON jobs(created_at);

CREATE INDEX idx_audit_logs_org_id ON audit_logs(org_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_job_id ON audit_logs(job_id);
CREATE INDEX idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

CREATE INDEX idx_scan_results_job_id ON scan_results(job_id);
CREATE INDEX idx_scan_results_scan_type ON scan_results(scan_type);
CREATE INDEX idx_scan_results_created_at ON scan_results(created_at);
```

---

## <a name="api"></a>5. API Design

### 5.1 REST API

**Base URL**: `https://api.warden-cicd.dev/v1`

**Authentication**: Bearer token (JWT) in Authorization header

**Request Headers**:
```
Authorization: Bearer <token>
Content-Type: application/json
X-Request-ID: <request-id>
```

**Response Format**:
```json
{
  "success": true,
  "data": { ... },
  "error": null
}

// Error response
{
  "success": false,
  "data": null,
  "error": {
    "code": "validation_error",
    "message": "Invalid policy YAML",
    "details": { ... }
  }
}
```

### 5.2 API Endpoints

#### Authentication

| Method | Endpoint | Description | Request | Response |
|--------|----------|-------------|---------|----------|
| POST | `/auth/login` | User login | `{email, password}` | `{token, expires_at}` |
| POST | `/auth/refresh` | Refresh token | `{refresh_token}` | `{token, expires_at}` |
| POST | `/auth/logout` | User logout | `{token}` | `{success}` |
| GET | `/auth/me` | Get current user | - | `{user}` |

#### Organizations

| Method | Endpoint | Description | Request | Response |
|--------|----------|-------------|---------|----------|
| GET | `/organizations` | List organizations | - | `[{org}]` |
| POST | `/organizations` | Create organization | `{name}` | `{org}` |
| GET | `/organizations/:id` | Get organization | - | `{org}` |
| PUT | `/organizations/:id` | Update organization | `{name}` | `{org}` |
| DELETE | `/organizations/:id` | Delete organization | - | `{success}` |

#### Policies

| Method | Endpoint | Description | Request | Response |
|--------|----------|-------------|---------|----------|
| GET | `/policies` | List policies | - | `[{policy}]` |
| POST | `/policies` | Create policy | `{name, yaml}` | `{policy}` |
| GET | `/policies/:id` | Get policy | - | `{policy}` |
| PUT | `/policies/:id` | Update policy | `{name, yaml}` | `{policy}` |
| DELETE | `/policies/:id` | Delete policy | - | `{success}` |
| POST | `/policies/:id/test` | Test policy | `{command}` | `{result}` |
| GET | `/policies/templates` | List templates | - | `[{template}]` |

#### Jobs

| Method | Endpoint | Description | Request | Response |
|--------|----------|-------------|---------|----------|
| GET | `/jobs` | List jobs | `{org_id?, status?, limit?, offset?}` | `{jobs, total}` |
| POST | `/jobs` | Create job | `{policy_id, command}` | `{job}` |
| GET | `/jobs/:id` | Get job | - | `{job}` |
| GET | `/jobs/:id/logs` | Get job logs | - | `{logs}` |
| GET | `/jobs/:id/artifacts` | Get artifacts | - | `{url}` |
| DELETE | `/jobs/:id` | Cancel job | - | `{success}` |

#### Scanning

| Method | Endpoint | Description | Request | Response |
|--------|----------|-------------|---------|----------|
| POST | `/scan/sast` | Run SAST scan | `{job_id}` | `{scan}` |
| POST | `/scan/secrets` | Run secrets scan | `{job_id}` | `{scan}` |
| GET | `/scan/results` | List scan results | `{job_id?}` | `{results}` |
| GET | `/scan/results/:id` | Get scan result | - | `{result}` |

#### Audit

| Method | Endpoint | Description | Request | Response |
|--------|----------|-------------|---------|----------|
| GET | `/audit/logs` | List audit logs | `{org_id?, event_type?, limit?, offset?}` | `{logs, total}` |
| GET | `/audit/logs/export` | Export audit logs | `{format, filter}` | `{url}` |

#### Alerts

| Method | Endpoint | Description | Request | Response |
|--------|----------|-------------|---------|----------|
| GET | `/alerts` | List alerts | - | `[{alert}]` |
| POST | `/alerts` | Create alert | `{name, type, config}` | `{alert}` |
| GET | `/alerts/:id` | Get alert | - | `{alert}` |
| PUT | `/alerts/:id` | Update alert | `{name, type, config}` | `{alert}` |
| DELETE | `/alerts/:id` | Delete alert | - | `{success}` |

---

## <a name="integrations"></a>6. Integration Points

### 6.1 GitHub Actions Integration

#### Option 1: Warden GitHub Action

**Workflow**:
```yaml
name: Secure CI with Warden

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Run Build in Warden Sandbox
        uses: warden-dev/warden-action@v1
        with:
          policy: |
            version: "1.0"
            filesystem:
              read:
                - "/workspace/src"
              write:
                - "/workspace/build"
            network:
              allow:
                - "registry.npmjs.org"
          command: "npm run build"
```

**Implementation**:
- **Action**: Docker container with Warden CLI
- **Inputs**: Policy (YAML or file), Command
- **Outputs**: Exit code, stdout, stderr, logs

#### Option 2: Warden GitHub Actions Runner

**Setup**:
```bash
# Install Warden runner
curl -fsSL https://warden-cicd.dev/install | sh

# Configure runner
warden-cicd runner configure --token <github-token> --policy default

# Start runner
warden-cicd runner start
```

**Workflow**:
```yaml
name: Secure CI with Warden Runner

on: [push, pull_request]

jobs:
  build:
    runs-on: warden-runner
    steps:
      - uses: actions/checkout@v4
      - run: npm run build
```

**Implementation**:
- **Runner**: Custom GitHub Actions runner with Warden integration
- **Configuration**: Policy file or ID
- **Execution**: All jobs run in Warden sandboxes

### 6.2 Warden Core Integration

**Integration Points**:

1. **Sandbox Creation**
```go
import "warden-starter/warden"

func CreateSandbox(policy *Policy) (*warden.Sandbox, error) {
    config := convertPolicy(policy)
    return warden.NewSandbox(config)
}
```

2. **Policy Application**
```go
func ApplyPolicy(sandbox *warden.Sandbox, policy *Policy) error {
    // Convert and apply policy
    return sandbox.ApplyConfig(convertPolicy(policy))
}
```

3. **Command Execution**
```go
func Execute(sandbox *warden.Sandbox, command string) (*exec.Cmd, error) {
    return sandbox.Exec("sh", "-c", command)
}
```

4. **Resource Monitoring**
```go
func Monitor(sandbox *warden.Sandbox) (<-chan warden.Event, error) {
    return sandbox.Monitor()
}
```

5. **Audit Logging**
```go
func EnableAudit(sandbox *warden.Sandbox, handler func(event warden.Event)) error {
    return sandbox.EnableAudit(handler)
}
```

### 6.3 Scanning Integration

#### Semgrep Integration

**Installation**:
```bash
# Install Semgrep
pip install semgrep

# Or use Docker
docker pull returntocorp/semgrep
```

**Usage in Sandbox**:
```go
func RunSemgrep(sandbox *warden.Sandbox, dir string) ([]Finding, error) {
    cmd := sandbox.Exec("semgrep", "scan", dir, "--config", "auto", "--json")
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    return parseSemgrepOutput(output)
}
```

**Custom Rules**:
```yaml
# .semgrep.yml
rules:
  - id: hardcoded-password
    pattern: "password.*=.*['\"][^'\"]+['\"]"
    message: "Hardcoded password detected"
    severity: ERROR
    languages: [javascript, python, go]
```

#### Secrets Detection

**Implementation**:
```go
package secrets

import (
    "regexp"
    "strings"
)

var Patterns = []Pattern{
    {
        ID:       "aws_access_key_id",
        Name:     "AWS Access Key ID",
        Regex:    regexp.MustCompile(`(A3T[A-Z0-9]|AKIA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16}`),
        Severity: "critical",
    },
    {
        ID:       "github_token",
        Name:     "GitHub Personal Access Token",
        Regex:    regexp.MustCompile(`(ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{36,}`),
        Severity: "critical",
    },
    // ... more patterns
}

type Pattern struct {
    ID       string
    Name     string
    Regex    *regexp.Regexp
    Severity string
}

func ScanFile(content string, filename string) []Finding {
    var findings []Finding
    
    for _, pattern := range Patterns {
        matches := pattern.Regex.FindAllStringSubmatch(content, -1)
        for _, match := range matches {
            findings = append(findings, Finding{
                Pattern:   pattern,
                File:      filename,
                Line:      findLineNumber(content, match[0]),
                Match:     match[0],
            })
        }
    }
    
    return findings
}
```

---

## <a name="security"></a>7. Security Model

### 7.1 Zero-Trust Architecture

**Principles**:
1. **Never Trust, Always Verify**
   - All access attempts are verified against policies
   - No implicit trust for any process or user

2. **Deny by Default**
   - All access is denied unless explicitly allowed
   - No exceptions or backdoors

3. **Least Privilege**
   - Each process gets minimum required access
   - Access is temporary and revocable

4. **Micro-Segmentation**
   - Each job runs in its own isolated sandbox
   - No shared resources between jobs

5. **Continuous Verification**
   - Every access attempt is logged and audited
   - Anomalies are detected and flagged

### 7.2 Access Control

#### Filesystem Access
- **Read**: Only explicitly allowed paths can be read
- **Write**: Only explicitly allowed paths can be written
- **Deny**: Explicitly denied paths are always blocked
- **Default**: All other paths are inaccessible

**Example Policy**:
```yaml
filesystem:
  read:
    - "/workspace/src"
    - "/workspace/.github"
  write:
    - "/workspace/build"
  deny:
    - "/etc"
    - "/home"
    - "~/.ssh"
    - "~/.aws"
```

#### Network Access
- **Allow**: Only explicitly allowed hosts can be connected to
- **Deny**: Explicitly denied hosts are always blocked
- **DNS Filtering**: DNS requests for non-allowed hosts are blocked
- **Default**: All other network access is denied

**Example Policy**:
```yaml
network:
  allow:
    - "api.github.com"
    - "registry.npmjs.org"
  deny:
    - "0.0.0.0/0"
```

#### Process Access
- **Allow**: Only explicitly allowed commands can be executed
- **Deny**: Explicitly denied commands are always blocked
- **Shell Restriction**: Shell access can be restricted
- **Default**: All other commands are denied

**Example Policy**:
```yaml
process:
  allow:
    - "/usr/bin/node"
    - "/usr/bin/npm"
    - "/usr/bin/git"
  deny:
    - "/bin/bash"
    - "/bin/sh"
    - "curl"
    - "wget"
```

#### Environment Access
- **Allow**: Only explicitly allowed environment variables are passed through
- **Deny**: Explicitly denied environment variables are filtered out
- **Redaction**: Sensitive values can be redacted in logs
- **Default**: All other environment variables are filtered out

**Example Policy**:
```yaml
env:
  allow:
    - "GITHUB_TOKEN"
    - "CI"
  deny:
    - "AWS_ACCESS_KEY_ID"
    - "AWS_SECRET_ACCESS_KEY"
    - "*_PASSWORD"
    - "*_SECRET"
```

### 7.3 Runtime Protection

#### Resource Limits
- **CPU Limits**: Limit CPU usage to prevent resource exhaustion
- **Memory Limits**: Limit memory usage to prevent OOM attacks
- **Timeout Limits**: Limit execution time to prevent hanging processes
- **Process Limits**: Limit number of processes to prevent fork bombs

**Example**:
```yaml
limits:
  memory_mb: 4096
  timeout_s: 3600
  cpu_cores: 2
```

#### Behavior Monitoring
- **Anomaly Detection**: Detect anomalous behavior (e.g., sudden network connections)
- **Rate Limiting**: Limit rate of operations (e.g., file reads, network requests)
- **Pattern Detection**: Detect suspicious patterns (e.g., data exfiltration)

### 7.4 Threat Model

#### Threat Actors
- **Opportunistic Attackers**: Low-Medium capability, financial gain motivation
- **Organized Crime**: Medium-High capability, financial gain motivation
- **Nation-State Actors**: High capability, espionage/disruption motivation
- **Insider Threats**: Medium-High capability, revenge/profit motivation
- **Hacktivists**: Low-Medium capability, protest/activism motivation

#### Attack Vectors
| Attack Vector | Description | Mitigation |
|---------------|-------------|------------|
| Supply Chain Attack | Compromise dependencies or build process | Sandboxing, dependency verification |
| Credential Theft | Steal credentials from developers or systems | Secrets detection, environment isolation |
| Code Injection | Inject malicious code into repositories | Pre-commit scanning, approval workflows |
| Dependency Hijacking | Compromise package registries or dependencies | Dependency pinning, integrity verification |
| Build System Compromise | Compromise CI/CD systems | Sandboxing, least privilege |
| Runtime Exploitation | Exploit vulnerabilities during execution | Sandboxing, runtime protection |
| Data Exfiltration | Steal sensitive data from CI/CD systems | Network access control, data loss prevention |

#### Assets at Risk
| Asset | Sensitivity | Protection Required |
|-------|-------------|---------------------|
| Source Code | High | Encryption, access control, audit logging |
| Secrets | Critical | Secrets detection, environment isolation, encryption |
| Build Artifacts | High | Integrity verification, access control, signing |
| Credentials | Critical | MFA, least privilege, audit logging |
| Configuration | Medium-High | Access control, audit logging, encryption |
| Audit Logs | Medium | Integrity verification, access control, encryption |

---

## <a name="deployment"></a>8. Deployment Architecture

### 8.1 Deployment Topology

```
┌─────────────────────────────────────────────────────────────────────┐
│                        PRODUCTION ENVIRONMENT                        │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────────────┐  │
│  │  Load      │    │  Web       │    │      CONTROL PLANE          │  │
│  │  Balancer  │────▶│  Servers   │────▶│                             │  │
│  │  (ALB)     │    │  (x3)      │    │  ┌───────────────────────┐  │  │
│  └─────────────┘    └─────────────┘    │  │   API Server        │  │  │
│                                            └──────────┬────────────┘  │  │
│                                                       │              │  │
│  ┌───────────────────────────────────────────────────▼──────────────┐  │  │
│  │                        PostgreSQL Cluster                        │  │  │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌───────────────────────┐  │  │
│  │  │ Primary │  │ Replica │  │ Replica │  │   Redis (Cache)     │  │  │
│  │  └─────────┘  └─────────┘  └─────────┘  └───────────────────────┘  │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                        DATA PLANE                               │  │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌───────────────────────┐  │  │
│  │  │ Worker  │  │ Worker  │  │ Worker  │  │   Object Storage    │  │  │
│  │  │  1      │  │  2      │  │  3      │  │   (S3)              │  │  │
│  │  └─────────┘  └─────────┘  └─────────┘  └───────────────────────┘  │  │
│  └─────────────────────────────────────────────────────────────────┘  │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                        MONITORING                               │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌───────────────────────────┐  │  │
│  │  │  Prometheus │  │  Grafana    │  │   AlertManager         │  │  │
│  │  └─────────────┘  └─────────────┘  └───────────────────────────┘  │  │
│  └─────────────────────────────────────────────────────────────────┘  │
│                                                                       │
└─────────────────────────────────────────────────────────────────────┘
```

### 8.2 Infrastructure Requirements

#### Development Environment
- **Servers**: 1-2 development servers
- **CPU**: 8-16 vCPUs
- **Memory**: 32-64 GB RAM
- **Storage**: 500 GB SSD
- **Cost**: $500-$1,000/month

#### Staging Environment
- **Servers**: 2-3 servers (API, workers)
- **CPU**: 16-32 vCPUs
- **Memory**: 64-128 GB RAM
- **Storage**: 1 TB SSD
- **Cost**: $2,000-$4,000/month

#### Production Environment
- **Load Balancer**: ALB/NLB (3 AZs)
- **Web Servers**: 3 x c6i.large (4 vCPU, 8 GB RAM)
- **API Servers**: 3 x c6i.xlarge (8 vCPU, 16 GB RAM)
- **Workers**: 5 x c6i.xlarge (8 vCPU, 16 GB RAM)
- **Database**: RDS PostgreSQL (db.t3.xlarge, Multi-AZ)
- **Cache**: ElastiCache Redis (cache.t3.medium)
- **Storage**: S3 (500 GB)
- **Monitoring**: Prometheus, Grafana, AlertManager
- **Cost**: $5,000-$10,000/month

### 8.3 Deployment Strategy

#### Phase 1: Development (Weeks 1-8)
- **Environment**: Local development + shared dev server
- **Deployment**: Manual deployment for testing
- **Focus**: Feature development and testing

#### Phase 2: Staging (Weeks 9-10)
- **Environment**: Staging environment (cloud-based)
- **Deployment**: Automated CI/CD pipeline
- **Focus**: Integration testing, beta testing

#### Phase 3: Production (Weeks 11-12)
- **Environment**: Production environment (cloud-based)
- **Deployment**: Automated CI/CD pipeline with rollback
- **Focus**: Production launch, monitoring, scaling

### 8.4 CI/CD Pipeline

**Pipeline Stages**:

1. **Lint & Format**
   - Run linters (golangci-lint)
   - Run formatters (gofmt)
   - Check code style

2. **Unit Tests**
   - Run all unit tests
   - Check code coverage (>80%)
   - Run race detector

3. **Integration Tests**
   - Run integration tests
   - Test database interactions
   - Test API endpoints

4. **Build**
   - Build binaries (Linux, macOS, Windows)
   - Build Docker images
   - Build frontend assets

5. **Security Tests**
   - Run SAST (Semgrep)
   - Run secrets detection
   - Run dependency scanning
   - Run container scanning (if applicable)

6. **Deploy to Staging**
   - Deploy to staging environment
   - Run smoke tests
   - Run end-to-end tests

7. **Deploy to Production**
   - Deploy to production (blue-green or canary)
   - Monitor for errors
   - Rollback if needed

**Pipeline Definition (GitHub Actions)**:
```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
      - run: go mod download
      - run: golangci-lint run

  test:
    runs-on: ubuntu-latest
    needs: lint
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
      - run: go mod download
      - run: go test -v -race -coverprofile=coverage.out ./...
      - run: go tool cover -func=coverage.out | grep "total:" | awk '{print $3}' | sed 's/%//' > coverage.txt
      - name: Check coverage
        run: |
          COVERAGE=$(cat coverage.txt)
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage is below 80%: $COVERAGE%"
            exit 1
          fi

  build:
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
      - run: go mod download
      - run: go build -o warden-cicd
      - uses: docker/build-push-action@v4
        with:
          context: .
          push: true
          tags: warden-cicd:latest

  security:
    runs-on: ubuntu-latest
    needs: build
    steps:
      - uses: actions/checkout@v4
      - uses: returntocorp/semgrep-action@v1
        with:
          config: p/ci

  deploy-staging:
    runs-on: ubuntu-latest
    needs: security
    if: github.ref == 'refs/heads/develop'
    steps:
      - uses: actions/checkout@v4
      - run: echo "Deploy to staging"
      # Add deployment commands

  deploy-production:
    runs-on: ubuntu-latest
    needs: deploy-staging
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - run: echo "Deploy to production"
      # Add deployment commands
```

---

## 🎯 Conclusion

This **technical architecture** provides a **solid foundation** for Warden CI/CD MVP:

✅ **Proven Technology**: Uses Warden core (already exists and works)
✅ **Scalable Design**: Built for horizontal scaling
✅ **Secure by Default**: Deny-by-default, zero-trust architecture
✅ **Developer-Friendly**: Easy to use and understand
✅ **Extensible**: Built for future growth

**Next Steps**:
1. ✅ Review this architecture with the team
2. ✅ Finalize technical decisions
3. ✅ Begin implementation (Week 3)

---

## 📂 Document Structure

This is **Part 2** of the MVP Implementation Plan. For other parts, see:

1. **01-overview.md** - Executive summary and high-level plan
2. **02-technical-architecture.md** - This document
3. **03-timeline.md** - Week-by-week detailed timeline
4. **04-team-resources.md** - Team structure and hiring plan
5. **05-budget.md** - Detailed budget breakdown
6. **06-features.md** - Detailed feature specifications
7. **07-risks.md** - Risk management plan
8. **08-metrics.md** - Success metrics and tracking
