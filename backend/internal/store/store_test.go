package store

import (
	"sync"
	"testing"

	"github.com/Prof-bilal/ShieldCI/backend/internal/policy"
)

func TestMemoryClaimNextQueuedIsAtomic(t *testing.T) {
	db := NewMemory()
	created, err := db.CreatePolicy(policy.Policy{Version: "v1", Name: "claim-fixture"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateJob(created.ID, []string{"/bin/true"}); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	claimed := 0
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := db.ClaimNextQueued(); err == nil {
				mu.Lock()
				claimed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if claimed != 1 {
		t.Fatalf("expected one claimant, got %d", claimed)
	}
}
