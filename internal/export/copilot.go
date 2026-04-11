package export

import (
	"fmt"
	"strings"

	"github.com/arithmetike/ktext/internal/schema"
)

func renderCopilot(doc *schema.Context) string {
	var b strings.Builder
	id := doc.Identity
	w := working(doc)

	fmt.Fprintf(&b, "# Copilot Instructions — %s\n\n", id.Name)
	if id.Purpose != "" {
		fmt.Fprintf(&b, "%s\n\n", id.Purpose)
	}

	var instructions []string
	for _, c := range doc.Constraints {
		instructions = append(instructions, c.Content)
	}
	for _, c := range doc.Conventions {
		instructions = append(instructions, c.Rule)
	}
	if len(instructions) > 0 {
		b.WriteString("## Instructions\n\n")
		for _, r := range instructions {
			fmt.Fprintf(&b, "- %s\n", r)
		}
		b.WriteByte('\n')
	}

	if len(w.Commands) > 0 {
		b.WriteString("## Development Commands\n\n")
		for _, c := range w.Commands {
			fmt.Fprintf(&b, "- **%s**: `%s`\n", c.Description, c.Command)
		}
		b.WriteByte('\n')
	}

	if len(w.Structure) > 0 {
		b.WriteString("## Project Structure\n\n")
		for _, s := range w.Structure {
			fmt.Fprintf(&b, "- `%s` — %s\n", s.Path, s.Description)
		}
		b.WriteByte('\n')
	}

	if len(doc.Decisions) > 0 {
		b.WriteString("## Architecture Context\n\n")
		for _, d := range doc.Decisions {
			if d.Rationale != "" {
				fmt.Fprintf(&b, "- **%s**: %s\n", d.Title, d.Rationale)
			} else {
				fmt.Fprintf(&b, "- %s\n", d.Title)
			}
		}
		b.WriteByte('\n')
	}

	return b.String()
}
