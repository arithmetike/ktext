package export_test

import (
	"encoding/json"
	"testing"

	"github.com/arithmetike/ktext/internal/export"
	"github.com/arithmetike/ktext/internal/schema"
)

func TestJSONExportUsesLowercaseKeys(t *testing.T) {
	doc := &schema.Context{
		Ktext: "1.0.0",
		Identity: schema.Identity{
			Name:    "demo",
			URL:     "https://example.test/demo",
			Type:    "tool",
			Purpose: "a demo project",
			Status:  "active",
		},
		Constraints: []schema.Constraint{
			{Content: "never do X", Scope: "all", Why: "safety"},
		},
		Decisions: []schema.Decision{
			{Title: "use Go", Rationale: "fast builds", Date: "2026-04", Status: "accepted"},
		},
		Conventions: []schema.Convention{
			{Rule: "use gofmt", Why: "consistency"},
		},
		Risks: []schema.Risk{
			{Content: "schema renames break files", Severity: "high", Mitigation: "version field"},
		},
		Dependencies: []schema.Dependency{
			{Name: "yaml.v3", URL: "https://pkg.go.dev/gopkg.in/yaml.v3", Why: "YAML parsing"},
		},
		Working: &schema.Working{
			Commands: []schema.Command{
				{Command: "go build ./...", Description: "build"},
			},
			Structure: []schema.StructureEntry{
				{Path: "cmd/", Description: "CLI entry"},
			},
			Notes: []string{"run tests first"},
		},
		Ownership: &schema.Ownership{
			Team:       "platform",
			Escalation: "#oncall",
			Maintainers: []schema.Maintainer{
				{Name: "alice", Role: "lead"},
			},
		},
	}

	out, err := export.Render("json", doc)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Top-level keys must be lowercase schema names.
	for _, key := range []string{"ktext", "identity", "constraints", "decisions", "conventions", "risks", "dependencies", "working", "ownership"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing top-level key %q", key)
		}
	}

	// Nested identity keys must be lowercase.
	identity, ok := raw["identity"].(map[string]any)
	if !ok {
		t.Fatal("identity is not an object")
	}
	for _, key := range []string{"name", "url", "type", "purpose", "status"} {
		if _, ok := identity[key]; !ok {
			t.Errorf("identity: missing key %q", key)
		}
	}

	// Go field names must NOT appear.
	for _, bad := range []string{"Name", "URL", "Type", "Purpose", "Status", "Content", "Severity", "Mitigation", "Title", "Rationale", "Rule", "Command", "Description", "Path", "Team", "Escalation", "Maintainers"} {
		if _, ok := identity[bad]; ok {
			t.Errorf("identity: Go field name %q should not appear", bad)
		}
	}

	// Check nested struct keys in constraints.
	constraints, ok := raw["constraints"].([]any)
	if !ok || len(constraints) == 0 {
		t.Fatal("constraints is not a non-empty array")
	}
	c0, ok := constraints[0].(map[string]any)
	if !ok {
		t.Fatal("constraints[0] is not an object")
	}
	for _, key := range []string{"content", "scope", "why"} {
		if _, ok := c0[key]; !ok {
			t.Errorf("constraints[0]: missing key %q", key)
		}
	}

	// Check working.commands nested keys.
	working, ok := raw["working"].(map[string]any)
	if !ok {
		t.Fatal("working is not an object")
	}
	cmds, ok := working["commands"].([]any)
	if !ok || len(cmds) == 0 {
		t.Fatal("working.commands is not a non-empty array")
	}
	cmd0, ok := cmds[0].(map[string]any)
	if !ok {
		t.Fatal("working.commands[0] is not an object")
	}
	for _, key := range []string{"command", "description"} {
		if _, ok := cmd0[key]; !ok {
			t.Errorf("working.commands[0]: missing key %q", key)
		}
	}

	// Check ownership.maintainers nested keys.
	ownership, ok := raw["ownership"].(map[string]any)
	if !ok {
		t.Fatal("ownership is not an object")
	}
	maintainers, ok := ownership["maintainers"].([]any)
	if !ok || len(maintainers) == 0 {
		t.Fatal("ownership.maintainers is not a non-empty array")
	}
	m0, ok := maintainers[0].(map[string]any)
	if !ok {
		t.Fatal("ownership.maintainers[0] is not an object")
	}
	for _, key := range []string{"name", "role"} {
		if _, ok := m0[key]; !ok {
			t.Errorf("ownership.maintainers[0]: missing key %q", key)
		}
	}
}
