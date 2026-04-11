package export

import (
	"encoding/json"

	"github.com/thegagne/ktext/internal/schema"
)

func renderJSON(doc *schema.Context) (string, error) {
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil
}
