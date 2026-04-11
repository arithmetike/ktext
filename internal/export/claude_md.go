package export

import (
	"fmt"
	"strings"

	"github.com/thegagne/ktext/internal/schema"
)

func renderClaudeMd(doc *schema.Context) string {
	var b strings.Builder
	id := doc.Identity
	w := working(doc)

	fmt.Fprintf(&b, "# %s\n\n", id.Name)
	if id.Purpose != "" {
		fmt.Fprintf(&b, "%s\n\n", id.Purpose)
	}

	if len(w.Commands) > 0 {
		b.WriteString("## Commands\n\n")
		for _, c := range w.Commands {
			fmt.Fprintf(&b, "- %s: `%s`\n", capitalize(c.Description), c.Command)
		}
		b.WriteByte('\n')
	}

	if len(w.Structure) > 0 {
		b.WriteString("## Structure\n\n")
		for _, s := range w.Structure {
			fmt.Fprintf(&b, "- `%s` — %s\n", s.Path, s.Description)
		}
		b.WriteByte('\n')
	}

	// Rules = constraints + conventions merged
	var rules []string
	for _, c := range doc.Constraints {
		if c.Why != "" {
			rules = append(rules, c.Content+" ("+c.Why+")")
		} else {
			rules = append(rules, c.Content)
		}
	}
	for _, c := range doc.Conventions {
		if c.Why != "" {
			rules = append(rules, c.Rule+" ("+c.Why+")")
		} else {
			rules = append(rules, c.Rule)
		}
	}
	if len(rules) > 0 {
		b.WriteString("## Rules\n\n")
		for _, r := range rules {
			fmt.Fprintf(&b, "- %s\n", r)
		}
		b.WriteByte('\n')
	}

	if len(doc.Decisions) > 0 {
		b.WriteString("## Decisions\n\n")
		for _, d := range doc.Decisions {
			if d.Rationale != "" {
				fmt.Fprintf(&b, "- %s: %s\n", d.Title, d.Rationale)
			} else {
				fmt.Fprintf(&b, "- %s\n", d.Title)
			}
		}
		b.WriteByte('\n')
	}

	if len(w.Notes) > 0 {
		b.WriteString("## Notes\n\n")
		for _, n := range w.Notes {
			fmt.Fprintf(&b, "- %s\n", n)
		}
		b.WriteByte('\n')
	}

	if len(doc.Risks) > 0 {
		b.WriteString("## Risks\n\n")
		for _, r := range doc.Risks {
			if r.Mitigation != "" {
				fmt.Fprintf(&b, "- [%s] %s — Mitigation: %s\n", r.Severity, r.Content, r.Mitigation)
			} else {
				fmt.Fprintf(&b, "- [%s] %s\n", r.Severity, r.Content)
			}
		}
		b.WriteByte('\n')
	}

	return b.String()
}
