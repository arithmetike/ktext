package export

import (
	"fmt"
	"strings"

	"github.com/arithmetike/ktext/internal/schema"
)

func renderCursorrules(doc *schema.Context) string {
	var b strings.Builder
	id := doc.Identity
	w := working(doc)

	fmt.Fprintf(&b, "Project: %s\n", id.Name)
	if id.Type != "" {
		fmt.Fprintf(&b, "Type: %s\n", id.Type)
	}
	if id.Purpose != "" {
		fmt.Fprintf(&b, "Purpose: %s\n", id.Purpose)
	}
	b.WriteByte('\n')

	var rules []string
	for _, c := range doc.Constraints {
		rules = append(rules, c.Content)
	}
	for _, c := range doc.Conventions {
		rules = append(rules, c.Rule)
	}
	if len(rules) > 0 {
		b.WriteString("Rules:\n")
		for _, r := range rules {
			fmt.Fprintf(&b, "- %s\n", r)
		}
		b.WriteByte('\n')
	}

	if len(w.Commands) > 0 {
		b.WriteString("Commands:\n")
		for _, c := range w.Commands {
			fmt.Fprintf(&b, "- %s: %s\n", c.Description, c.Command)
		}
		b.WriteByte('\n')
	}

	if len(w.Structure) > 0 {
		b.WriteString("Structure:\n")
		for _, s := range w.Structure {
			fmt.Fprintf(&b, "- %s — %s\n", s.Path, s.Description)
		}
		b.WriteByte('\n')
	}

	if len(doc.Decisions) > 0 {
		b.WriteString("Key decisions:\n")
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
		b.WriteString("Notes:\n")
		for _, n := range w.Notes {
			fmt.Fprintf(&b, "- %s\n", n)
		}
		b.WriteByte('\n')
	}

	return b.String()
}
