package export

import (
	"github.com/arithmetike/ktext/internal/schema"
)

// working returns doc.Working, substituting an empty struct when nil so
// renderers don't need to guard every field access.
func working(doc *schema.Context) *schema.Working {
	if doc.Working != nil {
		return doc.Working
	}
	return &schema.Working{}
}
