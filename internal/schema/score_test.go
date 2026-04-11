package schema_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/arithmetike/ktext/internal/schema"
)


var minimal = schema.Context{
	Ktext:    "1.0.0",
	Identity: schema.Identity{Name: "demo", URL: "https://example.test/demo"},
}

func clone(c schema.Context) schema.Context {
	// Deep-copy via marshal round-trip is overkill; shallow clone is enough for
	// the test fixtures here since we assign new slices rather than appending.
	return c
}

func TestEmptyDocSurfacesFixesForEverySect(t *testing.T) {
	r := schema.ScoreDocument(&minimal)
	if r.Score >= 20 {
		t.Errorf("expected low score, got %d", r.Score)
	}
	if len(r.Fixes) < 6 {
		t.Errorf("expected >=6 fixes, got %d", len(r.Fixes))
	}
	for _, name := range []string{"constraints", "decisions", "conventions", "risks", "dependencies", "working", "ownership"} {
		var sec *schema.SectionResult
		for i := range r.Sections {
			if r.Sections[i].Name == name {
				sec = &r.Sections[i]
				break
			}
		}
		if sec == nil {
			t.Fatalf("section %q not found", name)
		}
		if sec.Score != 0 {
			t.Errorf("%s: expected score 0, got %d", name, sec.Score)
		}
		if sec.Status != "fail" {
			t.Errorf("%s: expected status 'fail', got %q", name, sec.Status)
		}
	}
}

func TestIdentityScoresPartialWhenNameURLOnly(t *testing.T) {
	r := schema.ScoreDocument(&minimal)
	sec := findSection(t, r, "identity")
	if sec.Score != 5 {
		t.Errorf("expected identity score 5, got %d", sec.Score)
	}
}

func TestFullyPopulatedIdentityEarnsFullPoints(t *testing.T) {
	doc := clone(minimal)
	doc.Identity.Type = "tool"
	doc.Identity.Status = "active"
	doc.Identity.Purpose = "A small command line tool that generates structured context files for AI coding agents and humans"
	r := schema.ScoreDocument(&doc)
	sec := findSection(t, r, "identity")
	if sec.Score != sec.Max {
		t.Errorf("expected identity score %d, got %d", sec.Max, sec.Score)
	}
}

func TestHighSeverityRiskWithoutMitigationProducesFix(t *testing.T) {
	doc := clone(minimal)
	doc.Risks = []schema.Risk{
		{Content: "SQLite WAL corruption when database file deleted while server holds it", Severity: "high"},
	}
	r := schema.ScoreDocument(&doc)
	found := false
	for _, f := range r.Fixes {
		if strings.Contains(f, "mitigation") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected mitigation fix, got fixes: %v", r.Fixes)
	}
}

func TestDependencyWithoutWhyProducesFix(t *testing.T) {
	doc := clone(minimal)
	doc.Dependencies = []schema.Dependency{
		{Name: "postgres", URL: "https://pg.test", Relationship: "depends_on"},
	}
	r := schema.ScoreDocument(&doc)
	found := false
	for _, f := range r.Fixes {
		if strings.Contains(f, "dependencies[0]") && strings.Contains(f, "why") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected dependencies[0] why fix, got fixes: %v", r.Fixes)
	}
}

func TestScoringIsDeterministic(t *testing.T) {
	a := schema.ScoreDocument(&minimal)
	b := schema.ScoreDocument(&minimal)
	if !reflect.DeepEqual(a, b) {
		t.Error("scoring is not deterministic")
	}
}

func TestPopulatingSectionsRaisesScore(t *testing.T) {
	base := schema.ScoreDocument(&minimal).Score
	doc := clone(minimal)
	doc.Constraints = []schema.Constraint{
		{Content: "Never log raw card data or PII anywhere in the request pipeline"},
	}
	doc.Decisions = []schema.Decision{
		{Title: "Use PostgreSQL", Rationale: "Need transactional integrity, JSONB support, and mature operational tooling for this team's workload"},
	}
	doc.Ownership = &schema.Ownership{Team: "platform", Escalation: "#oncall"}
	next := schema.ScoreDocument(&doc).Score
	if next <= base {
		t.Errorf("expected score to rise: base=%d next=%d", base, next)
	}
}

func TestWorkingCommandsAndStructureEarnPoints(t *testing.T) {
	doc := clone(minimal)
	doc.Working = &schema.Working{
		Commands: []schema.Command{
			{Command: "go build ./...", Description: "build the project"},
			{Command: "go test ./...", Description: "run unit tests"},
			{Command: "go vet ./...", Description: "lint sources"},
		},
		Structure: []schema.StructureEntry{
			{Path: "cmd/", Description: "Source code"},
		},
		Notes: []string{"tsx required for tests"},
	}
	r := schema.ScoreDocument(&doc)
	sec := findSection(t, r, "working")
	if sec.Score != sec.Max {
		t.Errorf("expected working score %d, got %d", sec.Max, sec.Score)
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func findSection(t *testing.T, r schema.ScoreResult, name string) schema.SectionResult {
	t.Helper()
	for _, s := range r.Sections {
		if s.Name == name {
			return s
		}
	}
	t.Fatalf("section %q not found", name)
	return schema.SectionResult{}
}

