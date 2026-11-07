package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Prof-bilal/ShieldCI/backend/internal/policy"
	"github.com/Prof-bilal/ShieldCI/backend/internal/store"
	"github.com/Prof-bilal/ShieldCI/backend/internal/wardenadapter"
)

func fakeWarden(t *testing.T, exit string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "warden")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+exit+"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunNextPersistsPassAndAudit(t *testing.T) {
	db := store.NewMemory()
	policyRecord, err := db.CreatePolicy(policy.Policy{Version: "v1", Name: "runner-fixture"})
	if err != nil {
		t.Fatal(err)
	}
	job, err := db.CreateJob(policyRecord.ID, []string{"/bin/echo", "hello"})
	if err != nil {
		t.Fatal(err)
	}
	adapter, err := wardenadapter.New(wardenadapter.Config{BinaryPath: fakeWarden(t, "exit 0"), TempDir: t.TempDir(), WorkingDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	runner := Runner{Store: db, Adapter: adapter}
	claimed, err := runner.RunNext(context.Background())
	if err != nil || !claimed {
		t.Fatalf("run failed: claimed=%v err=%v", claimed, err)
	}
	finished, _ := db.GetJob(job.ID)
	if finished.Status != "passed" {
		t.Fatalf("unexpected status: %s", finished.Status)
	}
	events, _ := db.ListAudit(job.ID)
	if len(events) != 1 || events[0].Decision != "allowed" {
		t.Fatalf("unexpected audit: %+v", events)
	}
}

func TestRunNextMarksBlockedWardenExit(t *testing.T) {
	db := store.NewMemory()
	policyRecord, _ := db.CreatePolicy(policy.Policy{Version: "v1", Name: "blocked-fixture"})
	job, _ := db.CreateJob(policyRecord.ID, []string{"/bin/false"})
	adapter, _ := wardenadapter.New(wardenadapter.Config{BinaryPath: fakeWarden(t, "exit 23"), TempDir: t.TempDir(), WorkingDir: t.TempDir()})
	_, err := (Runner{Store: db, Adapter: adapter}).RunNext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	finished, _ := db.GetJob(job.ID)
	if finished.Status != "blocked" {
		t.Fatalf("unexpected status: %s", finished.Status)
	}
}
