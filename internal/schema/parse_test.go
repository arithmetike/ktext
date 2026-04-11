package schema_test

import (
	"testing"

	"github.com/thegagne/ktext/internal/schema"
)

func TestParsesMinimalValidDoc(t *testing.T) {
	r := schema.Parse([]byte("ktext: \"1.0.0\"\nidentity:\n  name: demo\n  url: https://example.test/demo\n"))
	if !r.OK() {
		t.Fatalf("expected ok, got errors: %v", r.Errors)
	}
	if r.Doc.Identity.Name != "demo" {
		t.Errorf("expected name 'demo', got %q", r.Doc.Identity.Name)
	}
}

func TestRejectsUnknownKtextVersion(t *testing.T) {
	r := schema.Parse([]byte("ktext: \"2.0.0\"\nidentity:\n  name: x\n  url: https://example.test/x\n"))
	if r.OK() {
		t.Fatal("expected error for unknown ktext version")
	}
}

func TestRejectsMissingIdentityURL(t *testing.T) {
	r := schema.Parse([]byte("ktext: \"1.0.0\"\nidentity:\n  name: demo\n"))
	if r.OK() {
		t.Fatal("expected error for missing identity.url")
	}
}

func TestRejectsEmptyOwnership(t *testing.T) {
	r := schema.Parse([]byte("ktext: \"1.0.0\"\nidentity:\n  name: x\n  url: https://example.test/x\nownership: {}\n"))
	if r.OK() {
		t.Fatal("expected error for empty ownership object")
	}
}

func TestRejectsOwnershipWithOnlyEscalation(t *testing.T) {
	r := schema.Parse([]byte("ktext: \"1.0.0\"\nidentity:\n  name: x\n  url: https://example.test/x\nownership:\n  escalation: '#oncall'\n"))
	if r.OK() {
		t.Fatal("expected error for ownership with only escalation")
	}
}

func TestAcceptsOwnershipWithTeamOnly(t *testing.T) {
	r := schema.Parse([]byte("ktext: \"1.0.0\"\nidentity:\n  name: x\n  url: https://example.test/x\nownership:\n  team: platform\n"))
	if !r.OK() {
		t.Fatalf("expected ok, got errors: %v", r.Errors)
	}
}

func TestRejectsDecisionMissingRationale(t *testing.T) {
	r := schema.Parse([]byte("ktext: \"1.0.0\"\nidentity:\n  name: x\n  url: https://example.test/x\ndecisions:\n  - title: pick db\n"))
	if r.OK() {
		t.Fatal("expected error for decision missing rationale")
	}
}
