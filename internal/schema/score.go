package schema

import (
	"fmt"
	"math"
	"strings"
)

// SectionResult holds the score for one top-level section of a CONTEXT.yaml.
type SectionResult struct {
	Name    string   `json:"name"`
	Score   int      `json:"score"`
	Max     int      `json:"max"`
	Status  string   `json:"status"` // "pass" | "warn" | "fail"
	Details []string `json:"details,omitempty"`
	Fixes   []string `json:"fixes,omitempty"`
}

// ScoreResult is the output of ScoreDocument.
type ScoreResult struct {
	Score    int
	Sections []SectionResult
	Fixes    []string
}

// ScoreDocument scores a parsed Context document. Total max is 100.
func ScoreDocument(doc *Context) ScoreResult {
	sections := []SectionResult{
		scoreIdentity(doc),
		scoreConstraints(doc),
		scoreDecisions(doc),
		scoreConventions(doc),
		scoreRisks(doc),
		scoreDependencies(doc),
		scoreWorking(doc),
		scoreOwnership(doc),
	}

	total := 0
	var fixes []string
	for _, s := range sections {
		total += s.Score
		fixes = append(fixes, s.Fixes...)
	}
	return ScoreResult{Score: total, Sections: sections, Fixes: fixes}
}

// ── Section scorers ──────────────────────────────────────────────────────────

func scoreIdentity(doc *Context) SectionResult {
	const max = 15
	id := doc.Identity
	score := 0
	var details, fixes []string

	score += 3 // name is required by schema
	details = append(details, "name present")

	if id.URL != "" {
		score += 2
		details = append(details, "url present")
	} else {
		fixes = append(fixes, "identity: add url (canonical repository URL)")
	}

	if id.Type != "" {
		score += 2
		details = append(details, "type: "+id.Type)
	} else {
		fixes = append(fixes, "identity: add type (service, library, tool, etc.)")
	}

	if id.Status != "" {
		score += 2
		details = append(details, "status: "+id.Status)
	} else {
		fixes = append(fixes, "identity: add status (active, deprecated, experimental, archived)")
	}

	if id.Purpose != "" {
		score += 3
		if issues := CheckQuality("identity.purpose", id.Purpose); len(issues) == 0 {
			score += 3
			details = append(details, "purpose is specific")
		} else {
			score += 1
			fixes = append(fixes, "identity.purpose: "+issues[0])
		}
	} else {
		fixes = append(fixes, "identity: add purpose (what this does and why)")
	}

	return section("identity", score, max, details, fixes)
}

func scoreConstraints(doc *Context) SectionResult {
	const max = 20
	entries := doc.Constraints
	if len(entries) == 0 {
		return section("constraints", 0, max, nil, []string{
			"constraints: add at least 1 constraint — what must never happen here?",
		})
	}

	score := 8
	var details, fixes []string
	details = append(details, fmt.Sprintf("%d entries", len(entries)))

	perEntry := 12.0 / math.Max(float64(len(entries)), 3)
	qualityScore := 0.0

	for i, e := range entries {
		if issues := CheckQuality("constraints", e.Content); len(issues) == 0 {
			qualityScore += perEntry
			details = append(details, fmt.Sprintf("[%d] quality ok", i))
		} else {
			qualityScore += perEntry * 0.3
			fixes = append(fixes, fmt.Sprintf(`constraints[%d]: %q → %s`, i, e.Content, issues[0]))
		}
		if e.Scope != "" {
			details = append(details, fmt.Sprintf("[%d] has scope", i))
		}
		if e.Why != "" {
			qualityScore += 0.5
			details = append(details, fmt.Sprintf("[%d] has why", i))
		}
	}

	score += min(int(math.Round(qualityScore)), 12)
	return section("constraints", min(score, max), max, details, fixes)
}

func scoreDecisions(doc *Context) SectionResult {
	const max = 15
	entries := doc.Decisions
	if len(entries) == 0 {
		return section("decisions", 0, max, nil, []string{
			"decisions: add at least 1 architectural decision with rationale",
		})
	}

	score := 5
	var details, fixes []string
	details = append(details, fmt.Sprintf("%d entries", len(entries)))

	perEntry := 10.0 / math.Max(float64(len(entries)), 3)
	qualityScore := 0.0

	for i, d := range entries {
		entryScore := 0.0
		if issues := CheckQuality("decisions.rationale", d.Rationale); len(issues) == 0 {
			entryScore = perEntry
		} else {
			entryScore = perEntry * 0.4
			fixes = append(fixes, fmt.Sprintf(`decisions[%d]: %q rationale: %q → %s`, i, d.Title, truncate(d.Rationale, 60), issues[0]))
		}
		if d.Status != "" {
			entryScore += 0.3
		}
		if d.Date != "" {
			entryScore += 0.2
		}
		qualityScore += entryScore
	}

	score += min(int(math.Round(qualityScore)), 10)
	return section("decisions", min(score, max), max, details, fixes)
}

func scoreConventions(doc *Context) SectionResult {
	const max = 15
	entries := doc.Conventions
	if len(entries) == 0 {
		return section("conventions", 0, max, nil, []string{
			"conventions: add at least 1 coding or process convention",
		})
	}

	score := 5
	var details, fixes []string
	details = append(details, fmt.Sprintf("%d entries", len(entries)))

	perEntry := 10.0 / math.Max(float64(len(entries)), 3)
	qualityScore := 0.0

	for i, c := range entries {
		if issues := CheckQuality("conventions", c.Rule); len(issues) == 0 {
			qualityScore += perEntry
		} else {
			qualityScore += perEntry * 0.4
			fixes = append(fixes, fmt.Sprintf(`conventions[%d]: %q → %s`, i, c.Rule, issues[0]))
		}
		if c.Why != "" || c.Reference != "" {
			qualityScore += 0.5
		}
	}

	score += min(int(math.Round(qualityScore)), 10)
	return section("conventions", min(score, max), max, details, fixes)
}

func scoreRisks(doc *Context) SectionResult {
	const max = 10
	entries := doc.Risks
	if len(entries) == 0 {
		return section("risks", 0, max, nil, []string{
			"risks: add at least 1 risk with severity — what's fragile here?",
		})
	}

	score := 4
	var details, fixes []string
	details = append(details, fmt.Sprintf("%d entries", len(entries)))

	perEntry := 6.0 / math.Max(float64(len(entries)), 2)
	qualityScore := 0.0

	for i, r := range entries {
		if issues := CheckQuality("risks", r.Content); len(issues) == 0 {
			qualityScore += perEntry
		} else {
			qualityScore += perEntry * 0.3
			fixes = append(fixes, fmt.Sprintf(`risks[%d]: %q → %s`, i, r.Content, issues[0]))
		}
		if r.Mitigation != "" {
			qualityScore += 0.5
		} else if r.Severity == "high" {
			qualityScore -= 0.5
			fixes = append(fixes, fmt.Sprintf(`risks[%d]: %q → add a mitigation strategy`, i, r.Content))
		}
	}

	score += min(int(math.Round(qualityScore)), 6)
	return section("risks", min(score, max), max, details, fixes)
}

func scoreDependencies(doc *Context) SectionResult {
	const max = 10
	entries := doc.Dependencies
	if len(entries) == 0 {
		return section("dependencies", 0, max, nil, []string{
			"dependencies: list what this project calls or depends on",
		})
	}

	score := 5
	var details, fixes []string
	details = append(details, fmt.Sprintf("%d entries", len(entries)))

	qualityScore := 0.0
	for i, d := range entries {
		qualityScore += 1
		if d.URL != "" {
			qualityScore += 0.5
		}
		if d.Why != "" {
			qualityScore += 0.5
		} else {
			fixes = append(fixes, fmt.Sprintf(`dependencies[%d]: %q → add why — what does this project use it for?`, i, d.Name))
		}
	}

	score += min(int(math.Round(qualityScore)), 5)
	return section("dependencies", min(score, max), max, details, fixes)
}

func scoreWorking(doc *Context) SectionResult {
	const max = 10
	w := doc.Working
	if w == nil {
		return section("working", 0, max, nil, []string{
			"working: add build/test commands, directory structure, and dev notes",
		})
	}

	score := 0
	var details, fixes []string

	if len(w.Commands) > 0 {
		score += 2
		details = append(details, fmt.Sprintf("%d commands", len(w.Commands)))

		haystack := strings.Builder{}
		for _, c := range w.Commands {
			haystack.WriteString(strings.ToLower(c.Description + " " + c.Command + " | "))
		}
		h := haystack.String()
		if strings.Contains(h, "build") {
			score++
		} else {
			fixes = append(fixes, "working.commands: add a build command")
		}
		if strings.Contains(h, "test") {
			score++
		} else {
			fixes = append(fixes, "working.commands: add a test command")
		}
		if len(w.Commands) >= 3 {
			score++
		}
	} else {
		fixes = append(fixes, "working.commands: add build, test, and other dev commands")
	}

	if len(w.Structure) > 0 {
		score += 2
		details = append(details, fmt.Sprintf("structure: %d entries", len(w.Structure)))

		described := 0
		for _, s := range w.Structure {
			if s.Description != "" && !strings.HasPrefix(strings.ToUpper(s.Description), "TODO") {
				described++
			}
		}
		if described > 0 {
			score++
			details = append(details, fmt.Sprintf("%d entries with descriptions", described))
		} else {
			fixes = append(fixes, "working.structure: replace TODO descriptions with real ones")
		}
	} else {
		fixes = append(fixes, "working.structure: add directory layout with descriptions")
	}

	if len(w.Notes) > 0 {
		score += 2
		details = append(details, fmt.Sprintf("%d notes", len(w.Notes)))
	} else {
		fixes = append(fixes, "working.notes: add dev gotchas, workflow tips, or known issues")
	}

	return section("working", min(score, max), max, details, fixes)
}

func scoreOwnership(doc *Context) SectionResult {
	const max = 5
	o := doc.Ownership
	if o == nil {
		return section("ownership", 0, max, nil, []string{
			"ownership: add team name, maintainers, or escalation contact",
		})
	}

	score := 0
	var details, fixes []string

	if o.Team != "" {
		score += 2
		details = append(details, "team: "+o.Team)
	} else {
		fixes = append(fixes, "ownership: add team name")
	}
	if len(o.Maintainers) > 0 {
		score++
		details = append(details, fmt.Sprintf("%d maintainers", len(o.Maintainers)))
	}
	if o.Escalation != "" {
		score += 2
		details = append(details, "escalation present")
	} else {
		fixes = append(fixes, "ownership: add escalation path (Slack channel, PagerDuty, email)")
	}

	return section("ownership", min(score, max), max, details, fixes)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func section(name string, score, max int, details, fixes []string) SectionResult {
	return SectionResult{
		Name:    name,
		Score:   score,
		Max:     max,
		Status:  statusFor(score, max),
		Details: details,
		Fixes:   fixes,
	}
}

func statusFor(score, max int) string {
	ratio := float64(score) / float64(max)
	switch {
	case ratio >= 0.7:
		return "pass"
	case ratio >= 0.3:
		return "warn"
	default:
		return "fail"
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

