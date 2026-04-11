package schema

import (
	"bytes"
	_ "embed"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

//go:embed context-yaml.schema.json
var schemaJSON []byte

var compiled *jsonschema.Schema

func init() {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaJSON))
	if err != nil {
		panic("ktext: failed to parse embedded schema: " + err.Error())
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource("context-yaml.schema.json", doc); err != nil {
		panic("ktext: failed to load embedded schema: " + err.Error())
	}
	compiled, err = c.Compile("context-yaml.schema.json")
	if err != nil {
		panic("ktext: failed to compile embedded schema: " + err.Error())
	}
}

// ValidationError is a single schema violation with a location and message.
type ValidationError struct {
	Path    []string // JSON Pointer segments, e.g. ["identity", "url"]
	Message string
}

// ParseResult is the outcome of parsing and validating a CONTEXT.yaml document.
type ParseResult struct {
	Doc    *Context
	Errors []ValidationError
}

// OK reports whether parsing and validation succeeded.
func (r ParseResult) OK() bool { return len(r.Errors) == 0 }

// Parse reads a CONTEXT.yaml document from src, validates it against the
// embedded JSON Schema, and returns the result.
func Parse(src []byte) ParseResult {
	// Decode YAML into a generic map first so the JSON Schema validator can
	// inspect it before we bind to strongly-typed structs.
	var raw any
	if err := yaml.Unmarshal(src, &raw); err != nil {
		return ParseResult{Errors: []ValidationError{{Message: "invalid YAML: " + err.Error()}}}
	}
	if raw == nil {
		return ParseResult{Errors: []ValidationError{{Message: "document is empty"}}}
	}

	// yaml.v3 unmarshals maps as map[string]any; jsonschema expects that too.
	if err := compiled.Validate(raw); err != nil {
		return ParseResult{Errors: formatErrors(err)}
	}

	// Second pass: bind into typed structs.
	var doc Context
	if err := yaml.Unmarshal(src, &doc); err != nil {
		return ParseResult{Errors: []ValidationError{{Message: "failed to decode document: " + err.Error()}}}
	}
	return ParseResult{Doc: &doc}
}

func formatErrors(err error) []ValidationError {
	var ve *jsonschema.ValidationError
	if !errorAs(err, &ve) {
		return []ValidationError{{Message: err.Error()}}
	}
	var out []ValidationError
	collectErrors(ve, &out)
	return out
}

func collectErrors(ve *jsonschema.ValidationError, out *[]ValidationError) {
	if len(ve.Causes) == 0 {
		*out = append(*out, ValidationError{
			Path:    ve.InstanceLocation,
			Message: ve.Error(),
		})
		return
	}
	for _, cause := range ve.Causes {
		collectErrors(cause, out)
	}
}

// errorAs is a thin wrapper around errors.As to avoid importing "errors" at
// the call site.
func errorAs(err error, target any) bool {
	type asIface interface{ As(any) bool }
	// Walk the error chain manually since we can't import errors here
	// without a cycle; the standard unwrap protocol suffices.
	for err != nil {
		if x, ok := err.(asIface); ok && x.As(target) {
			return true
		}
		// Try direct type assertion via the errors.As logic.
		if t, ok := target.(**jsonschema.ValidationError); ok {
			if v, ok := err.(*jsonschema.ValidationError); ok {
				*t = v
				return true
			}
		}
		type unwrapper interface{ Unwrap() error }
		if u, ok := err.(unwrapper); ok {
			err = u.Unwrap()
		} else {
			break
		}
	}
	return false
}
