package wardenadapter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Prof-bilal/ShieldCI/backend/internal/policy"
	"gopkg.in/yaml.v3"
)

var (
	ErrMissingBinary = errors.New("warden binary is not configured")
	ErrPolicyWrite   = errors.New("could not write generated Warden policy")
)

type Config struct {
	BinaryPath string
	Backend    string
	TempDir    string
	WorkingDir string
}

type AuditEvent struct {
	Type      string    `json:"type"`
	Decision  string    `json:"decision"`
	Reason    string    `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Result struct {
	ExitCode   int          `json:"exit_code"`
	Stdout     string       `json:"stdout"`
	Stderr     string       `json:"stderr"`
	StartedAt  time.Time    `json:"started_at"`
	FinishedAt time.Time    `json:"finished_at"`
	Events     []AuditEvent `json:"events"`
}

type Adapter struct {
	config Config
}

func New(config Config) (*Adapter, error) {
	if strings.TrimSpace(config.BinaryPath) == "" {
		return nil, ErrMissingBinary
	}
	if config.TempDir == "" {
		config.TempDir = os.TempDir()
	}
	if config.WorkingDir == "" {
		config.WorkingDir, _ = os.Getwd()
	}
	return &Adapter{config: config}, nil
}

func (a *Adapter) Run(ctx context.Context, p policy.Policy, command []string) (Result, error) {
	if err := p.Validate(); err != nil {
		return Result{}, fmt.Errorf("validate policy: %w", err)
	}
	if len(command) == 0 {
		return Result{}, errors.New("command is required")
	}
	if _, err := os.Stat(a.config.BinaryPath); err != nil {
		return Result{}, fmt.Errorf("warden binary unavailable: %w", err)
	}
	workingDir, err := filepath.Abs(a.config.WorkingDir)
	if err != nil {
		return Result{}, fmt.Errorf("resolve working directory: %w", err)
	}
	if info, err := os.Stat(workingDir); err != nil || !info.IsDir() {
		if err == nil {
			err = errors.New("not a directory")
		}
		return Result{}, fmt.Errorf("invalid working directory: %w", err)
	}

	policyFile, err := os.CreateTemp(a.config.TempDir, "shieldci-policy-*.yaml")
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrPolicyWrite, err)
	}
	policyPath := policyFile.Name()
	defer os.Remove(policyPath)
	defer policyFile.Close()

	wardenPolicy := struct {
		Filesystem policy.Filesystem  `yaml:"filesystem"`
		Network    policy.Network     `yaml:"network"`
		Env        policy.Environment `yaml:"env"`
	}{Filesystem: resolvePaths(p.Filesystem, workingDir), Network: p.Network, Env: p.Environment}
	encoded, err := yaml.Marshal(wardenPolicy)
	if err != nil {
		return Result{}, fmt.Errorf("%w: encode yaml: %v", ErrPolicyWrite, err)
	}
	if _, err := policyFile.Write(encoded); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrPolicyWrite, err)
	}
	if err := policyFile.Close(); err != nil {
		return Result{}, fmt.Errorf("%w: close policy: %v", ErrPolicyWrite, err)
	}

	args := []string{"run", "--policy", filepath.Clean(policyPath)}
	if a.config.Backend != "" {
		args = append(args, "--backend", a.config.Backend)
	}
	args = append(args, "--")
	args = append(args, command...)
	cmd := exec.CommandContext(ctx, a.config.BinaryPath, args...)
	cmd.Dir = workingDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	started := time.Now().UTC()
	err = cmd.Run()
	finished := time.Now().UTC()

	result := Result{Stdout: stdout.String(), Stderr: stderr.String(), StartedAt: started, FinishedAt: finished}
	event := AuditEvent{Type: "sandbox.run", Decision: "allowed", CreatedAt: finished}
	if err != nil {
		event.Decision = "blocked"
		event.Reason = "warden refused or command failed"
		result.Events = []AuditEvent{event}
		if ctx.Err() != nil {
			return result, fmt.Errorf("sandbox context: %w", ctx.Err())
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
			return result, nil
		}
		return result, fmt.Errorf("execute Warden: %w", err)
	}
	result.Events = []AuditEvent{event}
	return result, nil
}

func resolvePaths(filesystem policy.Filesystem, workingDir string) policy.Filesystem {
	resolve := func(paths []string) []string {
		resolved := make([]string, 0, len(paths))
		for _, path := range paths {
			if filepath.IsAbs(path) {
				resolved = append(resolved, filepath.Clean(path))
			} else {
				resolved = append(resolved, filepath.Join(workingDir, filepath.Clean(path)))
			}
		}
		return resolved
	}
	filesystem.Read = resolve(filesystem.Read)
	filesystem.Write = resolve(filesystem.Write)
	return filesystem
}
