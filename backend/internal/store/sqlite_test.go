package store

import (
	"testing"

	"github.com/Prof-bilal/ShieldCI/backend/internal/policy"
)

func TestSQLiteCLIRoundTrip(t *testing.T) {
	db, err := NewSQLite(t.TempDir() + "/shieldci.db")
	if err != nil {
		t.Fatal(err)
	}
	created, err := db.CreatePolicy(policy.Policy{Version: "v1", Name: "sqlite-fixture"})
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := db.GetPolicy(created.ID)
	if err != nil || loaded.Policy.Name != "sqlite-fixture" {
		t.Fatalf("unexpected policy: %+v, %v", loaded, err)
	}
	job, err := db.CreateJob(created.ID, []string{"/bin/echo", "hello"})
	if err != nil {
		t.Fatal(err)
	}
	loadedJob, err := db.GetJob(job.ID)
	if err != nil || loadedJob.Status != "queued" || len(loadedJob.Command) != 2 {
		t.Fatalf("unexpected job: %+v, %v", loadedJob, err)
	}
	claimed, err := db.ClaimNextQueued()
	if err != nil || claimed.ID != job.ID || claimed.Status != "running" {
		t.Fatalf("unexpected claim: %+v, %v", claimed, err)
	}
	if _, err := db.GetJob("missing"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
