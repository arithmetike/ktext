// Package export renders a parsed Context document into agent-specific formats.
package export

import (
	"fmt"

	"github.com/thegagne/ktext/internal/schema"
)

// Format describes a supported output format.
type Format struct {
	Name     string
	Filename string
	Desc     string
}

// All supported formats in display order.
var All = []Format{
	{"yaml", "CONTEXT.yaml", "Canonical CONTEXT.yaml (re-serialized)"},
	{"claude-md", "CLAUDE.md", "Claude Code project instructions"},
	{"cursorrules", ".cursorrules", "Cursor IDE rules"},
	{"copilot", ".github/copilot-instructions.md", "GitHub Copilot instructions"},
	{"xml", "context.xml", "Token-efficient XML for LLM injection"},
	{"json", "context.json", "Structured JSON"},
}

// Render converts doc to the named format. Returns an error for unknown formats.
func Render(format string, doc *schema.Context) (string, error) {
	switch format {
	case "yaml":
		return renderYAML(doc)
	case "claude-md":
		return renderClaudeMd(doc), nil
	case "cursorrules":
		return renderCursorrules(doc), nil
	case "copilot":
		return renderCopilot(doc), nil
	case "xml":
		return renderXML(doc), nil
	case "json":
		return renderJSON(doc)
	default:
		return "", fmt.Errorf("unknown format %q", format)
	}
}
