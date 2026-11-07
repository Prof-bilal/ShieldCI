package wardenadapter

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Prof-bilal/ShieldCI/backend/internal/policy"
)

func writeFakeWarden(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "warden")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func validPolicy() policy.Policy {
	return policy.Policy{Version: "v1", Name: "fixture", Filesystem: policy.Filesystem{Read: []string{"./workspace"}}}
}

func TestRunBuildsExplicitWardenCommand(t *testing.T) {
	workingDir := t.TempDir()
	bin := writeFakeWarden(t, `policy=""; previous=""; for arg in "$@"; do if [ "$previous" = "--policy" ]; then policy="$arg"; fi; previous="$arg"; done; printf '%s\n' "$*"; cat "$policy"; exit 0`)
	adapter, err := New(Config{BinaryPath: bin, TempDir: t.TempDir(), WorkingDir: workingDir, Backend: "linux"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := adapter.Run(context.Background(), validPolicy(), []string{"/bin/echo", "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Stdout, "run --policy") || !strings.Contains(result.Stdout, "--backend linux") {
		t.Fatalf("unexpected command: %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "filesystem:\n    read:\n        - "+filepath.Join(workingDir, "workspace")) || !strings.Contains(result.Stdout, "env:") {
		t.Fatalf("unexpected Warden policy: %q", result.Stdout)
	}
	if len(result.Events) != 1 || result.Events[0].Decision != "allowed" {
		t.Fatalf("unexpected audit events: %+v", result.Events)
	}
}

func TestRunReturnsExitCodeAndBlockedEvent(t *testing.T) {
	bin := writeFakeWarden(t, `echo denied >&2; exit 23`)
	adapter, err := New(Config{BinaryPath: bin, TempDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	result, err := adapter.Run(context.Background(), validPolicy(), []string{"/bin/false"})
	if err != nil {
		t.Fatal(err)
	}
	if result.ExitCode != 23 || result.Events[0].Decision != "blocked" || result.Stderr != "denied\n" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestRunFailsClosedWhenBinaryIsMissing(t *testing.T) {
	adapter, err := New(Config{BinaryPath: filepath.Join(t.TempDir(), "missing")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := adapter.Run(context.Background(), validPolicy(), []string{"/bin/true"}); err == nil {
		t.Fatal("expected missing Warden binary error")
	}
}
