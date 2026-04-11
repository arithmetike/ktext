// Package schema defines the CONTEXT.yaml v1 data model, parser, and validator.
//
// The canonical source of truth for the schema is context-yaml.schema.json.
// The Go types here mirror it; if you change one, change the other.
package schema

// Context is the top-level CONTEXT.yaml document.
type Context struct {
	Ktext        string         `yaml:"ktext"                  json:"ktext"`
	Identity     Identity       `yaml:"identity"               json:"identity"`
	Constraints  []Constraint   `yaml:"constraints,omitempty"  json:"constraints,omitempty"`
	Decisions    []Decision     `yaml:"decisions,omitempty"    json:"decisions,omitempty"`
	Conventions  []Convention   `yaml:"conventions,omitempty"  json:"conventions,omitempty"`
	Risks        []Risk         `yaml:"risks,omitempty"        json:"risks,omitempty"`
	Dependencies []Dependency   `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	Working      *Working       `yaml:"working,omitempty"      json:"working,omitempty"`
	Links        Links          `yaml:"links,omitempty"        json:"links,omitempty"`
	Ownership    *Ownership     `yaml:"ownership,omitempty"    json:"ownership,omitempty"`
	X            map[string]any `yaml:"x,omitempty"           json:"x,omitempty"`
}

// Identity describes what the project is.
type Identity struct {
	Name    string `yaml:"name"`
	URL     string `yaml:"url"`
	Type    string `yaml:"type,omitempty"`
	Purpose string `yaml:"purpose,omitempty"`
	Status  string `yaml:"status,omitempty"`
}

// Constraint is a rule engineers or AI must not violate.
type Constraint struct {
	Content string `yaml:"content"`
	Scope   string `yaml:"scope,omitempty"`
	Why     string `yaml:"why,omitempty"`
}

// Decision records an architectural choice and its rationale.
type Decision struct {
	Title     string `yaml:"title"`
	Rationale string `yaml:"rationale"`
	Date      string `yaml:"date,omitempty"`
	Status    string `yaml:"status,omitempty"`
	Reference string `yaml:"reference,omitempty"`
}

// Convention is a coding or process rule for the project.
type Convention struct {
	Rule      string `yaml:"rule"`
	Scope     string `yaml:"scope,omitempty"`
	Why       string `yaml:"why,omitempty"`
	Reference string `yaml:"reference,omitempty"`
}

// Risk is a known vulnerability, tech-debt item, or operational risk.
type Risk struct {
	Content    string `yaml:"content"`
	Severity   string `yaml:"severity"` // "high" | "medium" | "low"
	Mitigation string `yaml:"mitigation,omitempty"`
}

// Dependency is an outgoing connection from this project to another system.
type Dependency struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
	Why  string `yaml:"why,omitempty"`
}

// Working holds information a developer or LLM needs to operate in the codebase.
type Working struct {
	Commands  []Command        `yaml:"commands,omitempty"`
	Structure []StructureEntry `yaml:"structure,omitempty"`
	Notes     []string         `yaml:"notes,omitempty"`
}

// Command is a runnable shell command with a description.
type Command struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
}

// StructureEntry describes a path within the project.
type StructureEntry struct {
	Path        string `yaml:"path"`
	Description string `yaml:"description"`
}

// Links is an open map of name → URL.
type Links map[string]string

// Ownership records who maintains the project.
type Ownership struct {
	Team        string       `yaml:"team,omitempty"`
	Escalation  string       `yaml:"escalation,omitempty"`
	Maintainers []Maintainer `yaml:"maintainers,omitempty"`
}

// Maintainer is a named person responsible for the project.
type Maintainer struct {
	Name string `yaml:"name"`
	Role string `yaml:"role,omitempty"`
}
