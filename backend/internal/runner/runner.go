package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/Prof-bilal/ShieldCI/backend/internal/store"
	"github.com/Prof-bilal/ShieldCI/backend/internal/wardenadapter"
)

type Runner struct {
	Store   store.Store
	Adapter *wardenadapter.Adapter
}

// RunNext claims one queued job and always keeps execution behind Warden.
func (r Runner) RunNext(ctx context.Context) (bool, error) {
	if r.Store == nil || r.Adapter == nil {
		return false, fmt.Errorf("runner dependencies are required")
	}
	queued, err := r.Store.ClaimNextQueued()
	if err == store.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("claim job: %w", err)
	}
	record, err := r.Store.GetPolicy(queued.PolicyID)
	if err != nil {
		return true, r.finishFailure(queued.ID, "failed", "policy lookup failed: "+err.Error())
	}
	result, runErr := r.Adapter.Run(ctx, record.Policy, queued.Command)
	for _, event := range result.Events {
		if err := r.Store.AppendAudit(store.AuditEvent{JobID: queued.ID, Type: event.Type, Decision: event.Decision, Reason: event.Reason, CreatedAt: event.CreatedAt}); err != nil {
			return true, fmt.Errorf("record audit event: %w", err)
		}
	}
	status := "passed"
	if runErr != nil || result.ExitCode != 0 {
		status = "failed"
		if len(result.Events) > 0 && result.Events[0].Decision == "blocked" {
			status = "blocked"
		}
	}
	if _, err := r.Store.UpdateJobStatus(queued.ID, status); err != nil {
		return true, fmt.Errorf("finish job: %w", err)
	}
	return true, runErr
}

func (r Runner) finishFailure(jobID, status, reason string) error {
	if err := r.Store.AppendAudit(store.AuditEvent{JobID: jobID, Type: "runner.error", Decision: "blocked", Reason: reason, CreatedAt: time.Now().UTC()}); err != nil {
		return err
	}
	_, err := r.Store.UpdateJobStatus(jobID, status)
	return err
}
