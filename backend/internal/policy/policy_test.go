package policy

import (
	"errors"
	"testing"
)

func TestValidateAcceptsExplicitPolicy(t *testing.T) {
	p := Policy{Version: "v1", Name: "unit", Filesystem: Filesystem{Read: []string{"./workspace"}}, Network: Network{Allow: []string{"registry.npmjs.org"}}, Environment: Environment{Allow: []string{"CI"}}}
	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid policy, got %v", err)
	}
}

func TestValidateRejectsMissingIdentity(t *testing.T) {
	if err := (Policy{}).Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestValidateRejectsAmbiguousAccess(t *testing.T) {
	p := Policy{Version: "v1", Name: "bad", Filesystem: Filesystem{Read: []string{"~/.ssh"}}, Network: Network{Allow: []string{"https://evil.example"}}}
	if err := p.Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
