package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRunSemgrepParsesReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "semgrep")
	script := `#!/bin/sh
printf '%s' '{"results":[{"check_id":"python.lang.security","path":"app.py","start":{"line":7},"extra":{"message":"unsafe call","severity":"ERROR"}}]}'`
	if err := os.WriteFile(path, []byte(script+"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	findings, err := RunSemgrep(context.Background(), path, t.TempDir())
	if err != nil || len(findings) != 1 {
		t.Fatalf("unexpected semgrep findings: %+v, %v", findings, err)
	}
	if findings[0].Rule != "python.lang.security" || findings[0].Severity != "error" || findings[0].Line != 7 {
		t.Fatalf("unexpected finding: %+v", findings[0])
	}
}
