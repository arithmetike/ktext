package export

import (
	"github.com/thegagne/ktext/internal/schema"
	"gopkg.in/yaml.v3"
)

func renderYAML(doc *schema.Context) (string, error) {
	b, err := yaml.Marshal(doc)
	if err != nil {
		return "", err
	}
	header := "# CONTEXT.yaml — machine-readable context for this repository\n" +
		"# Schema: https://contextfile.org/schema/v1\n\n"
	return header + string(b), nil
}
