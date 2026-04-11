package export

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/thegagne/ktext/internal/schema"
)

func renderXML(doc *schema.Context) string {
	var b strings.Builder
	id := doc.Identity
	w := working(doc)

	attrs := fmt.Sprintf("name=%q", id.Name)
	if id.Type != "" {
		attrs += fmt.Sprintf(" type=%q", id.Type)
	}
	if id.Status != "" && id.Status != "active" {
		attrs += fmt.Sprintf(" status=%q", id.Status)
	}
	fmt.Fprintf(&b, "<context %s>\n", attrs)

	if id.Purpose != "" {
		fmt.Fprintf(&b, "  <purpose>%s</purpose>\n", xmlEsc(id.Purpose))
	}

	if len(doc.Constraints) > 0 {
		b.WriteString("  <constraints>\n")
		for _, c := range doc.Constraints {
			fmt.Fprintf(&b, "    <c>%s</c>\n", xmlEsc(c.Content))
		}
		b.WriteString("  </constraints>\n")
	}

	if len(doc.Conventions) > 0 {
		b.WriteString("  <conventions>\n")
		for _, c := range doc.Conventions {
			fmt.Fprintf(&b, "    <c>%s</c>\n", xmlEsc(c.Rule))
		}
		b.WriteString("  </conventions>\n")
	}

	if len(doc.Decisions) > 0 {
		b.WriteString("  <decisions>\n")
		for _, d := range doc.Decisions {
			if d.Rationale != "" {
				fmt.Fprintf(&b, "    <d title=%q>%s</d>\n", d.Title, xmlEsc(d.Rationale))
			} else {
				fmt.Fprintf(&b, "    <d title=%q/>\n", d.Title)
			}
		}
		b.WriteString("  </decisions>\n")
	}

	if len(doc.Risks) > 0 {
		b.WriteString("  <risks>\n")
		for _, r := range doc.Risks {
			fmt.Fprintf(&b, "    <r severity=%q>%s</r>\n", r.Severity, xmlEsc(r.Content))
		}
		b.WriteString("  </risks>\n")
	}

	if len(w.Commands) > 0 {
		b.WriteString("  <commands>\n")
		for _, c := range w.Commands {
			fmt.Fprintf(&b, "    <cmd desc=%q>%s</cmd>\n", c.Description, xmlEsc(c.Command))
		}
		b.WriteString("  </commands>\n")
	}

	if len(w.Structure) > 0 {
		b.WriteString("  <structure>\n")
		for _, s := range w.Structure {
			fmt.Fprintf(&b, "    <dir name=%q>%s</dir>\n", s.Path, xmlEsc(s.Description))
		}
		b.WriteString("  </structure>\n")
	}

	if len(doc.Dependencies) > 0 {
		b.WriteString("  <dependencies>\n")
		for _, d := range doc.Dependencies {
			if d.Why != "" {
				fmt.Fprintf(&b, "    <dep name=%q rel=%q>%s</dep>\n", d.Name, d.Relationship, xmlEsc(d.Why))
			} else {
				fmt.Fprintf(&b, "    <dep name=%q rel=%q/>\n", d.Name, d.Relationship)
			}
		}
		b.WriteString("  </dependencies>\n")
	}

	if len(w.Notes) > 0 {
		b.WriteString("  <notes>\n")
		for _, n := range w.Notes {
			fmt.Fprintf(&b, "    <n>%s</n>\n", xmlEsc(n))
		}
		b.WriteString("  </notes>\n")
	}

	b.WriteString("</context>\n")
	return b.String()
}

func xmlEsc(s string) string {
	var b strings.Builder
	xml.EscapeText(&b, []byte(s)) //nolint
	return b.String()
}
