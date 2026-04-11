// Package export renders a parsed Context document into agent-specific formats.
package export

import (
	"fmt"

	"github.com/arithmetike/ktext/internal/schema"
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
	{"xml", "context.xml", "Token-efficient XML for LLM injection"},
	{"json", "context.json", "Structured JSON"},
}

// Render converts doc to the named format. Returns an error for unknown formats.
func Render(format string, doc *schema.Context) (string, error) {
	switch format {
	case "yaml":
		return renderYAML(doc)
	case "xml":
		return renderXML(doc), nil
	case "json":
		return renderJSON(doc)
	default:
		return "", fmt.Errorf("unknown format %q", format)
	}
}
