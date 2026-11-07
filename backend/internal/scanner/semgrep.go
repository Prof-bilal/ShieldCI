package scanner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

var ErrSemgrepUnavailable = errors.New("semgrep binary is unavailable")

type semgrepResult struct {
	CheckID string `json:"check_id"`
	Path    string `json:"path"`
	Start   struct {
		Line int `json:"line"`
	} `json:"start"`
	Extra struct {
		Message  string `json:"message"`
		Severity string `json:"severity"`
	} `json:"extra"`
}

type semgrepReport struct {
	Results []semgrepResult `json:"results"`
}

func RunSemgrep(ctx context.Context, binary, root string) ([]Finding, error) {
	if strings.TrimSpace(binary) == "" {
		return nil, ErrSemgrepUnavailable
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve semgrep root: %w", err)
	}
	cmd := exec.CommandContext(ctx, binary, "--json", "--config", "auto", root)
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return nil, fmt.Errorf("semgrep failed: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("run semgrep: %w", err)
	}
	var report semgrepReport
	if err := json.Unmarshal(output, &report); err != nil {
		return nil, fmt.Errorf("decode semgrep report: %w", err)
	}
	findings := make([]Finding, 0, len(report.Results))
	for _, result := range report.Results {
		severity := strings.ToLower(result.Extra.Severity)
		if severity == "" {
			severity = "medium"
		}
		fingerprint := fmt.Sprintf("%s:%s:%d", result.CheckID, result.Path, result.Start.Line)
		findings = append(findings, Finding{ID: "finding_" + fingerprint, Rule: result.CheckID, Severity: severity, Path: result.Path, Line: result.Start.Line, Message: result.Extra.Message, Fingerprint: fingerprint})
	}
	return findings, nil
}
