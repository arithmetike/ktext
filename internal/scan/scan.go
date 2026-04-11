// Package scan discovers project metadata from the filesystem for ktext init.
package scan

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/arithmetike/ktext/internal/schema"
	"gopkg.in/yaml.v3"
)

// Result holds everything discovered from a repository scan.
type Result struct {
	Identity     schema.Identity
	Constraints  []schema.Constraint
	Decisions    []schema.Decision
	Conventions  []schema.Convention
	Dependencies []schema.Dependency
	Working      schema.Working
	Links        schema.Links
	Log          []string // human-readable discovery notes
}

// Repo scans the directory at repoPath and returns what it finds.
func Repo(repoPath string) Result {
	r := Result{Links: make(schema.Links)}

	r.Identity = scanIdentity(repoPath, &r.Log)
	r.Working = scanWorking(repoPath, &r.Log)
	r.Decisions = scanDecisions(repoPath, &r.Log)
	r.Conventions = scanConventions(repoPath, &r.Log)
	r.Constraints = scanConstraints(repoPath, &r.Log)
	r.Dependencies = scanDependencies(repoPath, &r.Log)
	r.Links = scanLinks(repoPath, r.Identity)
	return r
}

// ── Identity ──────────────────────────────────────────────────────────────────

func scanIdentity(repoPath string, log *[]string) schema.Identity {
	name := filepath.Base(repoPath)
	url := ""

	if raw, err := gitCmd(repoPath, "remote", "get-url", "origin"); err == nil {
		remote := strings.TrimSpace(raw)
		// Derive name from remote
		part := remote[strings.LastIndex(remote, "/")+1:]
		name = strings.TrimSuffix(part, ".git")
		// Normalise to HTTPS
		if strings.HasPrefix(remote, "git@") {
			// git@github.com:org/repo.git → https://github.com/org/repo
			url = regexp.MustCompile(`^git@([^:]+):`).ReplaceAllString(remote, "https://$1/")
			url = strings.TrimSuffix(url, ".git")
		} else if strings.HasPrefix(remote, "https://") || strings.HasPrefix(remote, "http://") {
			url = strings.TrimSuffix(remote, ".git")
		}
	}

	purpose := extractPurpose(repoPath)
	typ := detectType(repoPath)

	if url == "" {
		*log = append(*log, "⚠  No git remote — identity.url will need editing")
	} else {
		*log = append(*log, "✓  Identity: "+name+" ("+string(typ)+")")
	}
	if purpose != "" {
		*log = append(*log, "✓  Purpose from README")
	}

	id := schema.Identity{
		Name:   name,
		URL:    url,
		Type:   typ,
		Status: "active",
	}
	if url == "" {
		id.URL = "https://example.com/REPLACE-ME"
	}
	if purpose != "" {
		id.Purpose = purpose
	}
	return id
}

func extractPurpose(repoPath string) string {
	for _, name := range []string{"README.md", "readme.md", "README.rst", "README"} {
		b, err := os.ReadFile(filepath.Join(repoPath, name))
		if err != nil {
			continue
		}
		lines := strings.Split(string(b), "\n")
		pastTitle := false
		var collected []string
		for _, line := range lines {
			if !pastTitle {
				if strings.HasPrefix(line, "# ") {
					pastTitle = true
				}
				continue
			}
			trimmed := strings.TrimSpace(line)
			if len(collected) == 0 && (trimmed == "" || strings.HasPrefix(trimmed, "![") || strings.HasPrefix(trimmed, "[![")) {
				continue
			}
			if trimmed == "" && len(collected) > 0 {
				break
			}
			if strings.HasPrefix(trimmed, "#") {
				break
			}
			collected = append(collected, trimmed)
		}
		text := strings.Join(collected, " ")
		if len(text) >= 20 {
			return text
		}
	}
	return ""
}

func detectType(repoPath string) string {
	exists := func(rel string) bool {
		_, err := os.Stat(filepath.Join(repoPath, rel))
		return err == nil
	}

	if exists("Dockerfile") || exists("docker-compose.yml") || exists("docker-compose.yaml") {
		return "service"
	}

	if b, err := os.ReadFile(filepath.Join(repoPath, "package.json")); err == nil {
		var pkg map[string]any
		if json.Unmarshal(b, &pkg) == nil {
			if pkg["bin"] != nil {
				return "tool"
			}
			if pkg["main"] != nil || pkg["exports"] != nil || pkg["types"] != nil {
				return "library"
			}
		}
	}

	if exists("go.mod") {
		// cmd/ subdirectory signals a CLI / service
		if exists("cmd") {
			return "tool"
		}
	}

	if b, err := os.ReadFile(filepath.Join(repoPath, "Cargo.toml")); err == nil {
		s := string(b)
		if strings.Contains(s, "[[bin]]") {
			return "tool"
		}
		if strings.Contains(s, "[lib]") {
			return "library"
		}
	}

	return ""
}

// ── Working ───────────────────────────────────────────────────────────────────

func scanWorking(repoPath string, log *[]string) schema.Working {
	cmds := scanCommands(repoPath, log)
	dirs := scanStructure(repoPath)

	w := schema.Working{}
	if len(cmds) > 0 {
		w.Commands = cmds
	}
	if len(dirs) > 0 {
		w.Structure = dirs
		*log = append(*log, "✓  Directory structure ("+itoa(len(dirs))+" entries)")
	}
	return w
}

func scanCommands(repoPath string, log *[]string) []schema.Command {
	var cmds []schema.Command
	seen := map[string]bool{}
	add := func(desc, cmd string) {
		if seen[desc] {
			return
		}
		seen[desc] = true
		cmds = append(cmds, schema.Command{Command: cmd, Description: desc})
	}

	// package.json scripts
	if b, err := os.ReadFile(filepath.Join(repoPath, "package.json")); err == nil {
		var pkg struct {
			Scripts map[string]string `json:"scripts"`
		}
		if json.Unmarshal(b, &pkg) == nil {
			for _, name := range []string{"build", "test", "lint", "dev", "start", "format", "check", "typecheck", "type-check"} {
				if pkg.Scripts[name] != "" {
					add(name, "npm run "+name)
				}
			}
			if len(cmds) > 0 {
				*log = append(*log, "✓  Build commands from package.json ("+itoa(len(cmds))+" scripts)")
			}
		}
	}

	// Makefile targets
	if b, err := os.ReadFile(filepath.Join(repoPath, "Makefile")); err == nil {
		re := regexp.MustCompile(`(?m)^([\w-]+):`)
		interesting := map[string]bool{"build": true, "test": true, "lint": true, "dev": true, "run": true, "clean": true, "install": true, "deploy": true}
		for _, m := range re.FindAllSubmatch(b, -1) {
			t := string(m[1])
			if interesting[t] {
				add(t, "make "+t)
			}
		}
	}

	// Go module
	if _, err := os.Stat(filepath.Join(repoPath, "go.mod")); err == nil {
		add("build", "go build ./...")
		add("test", "go test ./...")
	}

	// Cargo
	if _, err := os.Stat(filepath.Join(repoPath, "Cargo.toml")); err == nil {
		add("build", "cargo build")
		add("test", "cargo test")
	}

	return cmds
}

var ignoreDirs = map[string]bool{
	"node_modules": true, ".git": true, "dist": true, "build": true,
	"vendor": true, ".next": true, "__pycache__": true, ".tox": true,
	"target": true, ".gradle": true, "coverage": true, ".turbo": true,
	".cache": true, "out": true, ".venv": true, "venv": true,
}

var knownDirs = map[string]string{
	"src": "Source code", "lib": "Library code", "app": "Application code",
	"pkg": "Reusable packages", "cmd": "CLI / service entry points",
	"internal": "Internal packages (not for external use)", "test": "Test suite",
	"tests": "Test suite", "__tests__": "Test suite", "spec": "Test suite",
	"e2e": "End-to-end tests", "docs": "Documentation", "doc": "Documentation",
	"scripts": "Build and dev scripts", "bin": "Executable scripts",
	"config": "Configuration files", "configs": "Configuration files",
	"examples": "Usage examples", "example": "Usage examples",
	"migrations": "Database migrations", "db": "Database schema and migrations",
	"api": "API surface and handlers", "web": "Web frontend",
	"ui": "User interface code", "public": "Static assets served directly",
	"static": "Static assets served directly", "assets": "Static assets",
	"templates": "Template files", "proto": "Protocol buffer definitions",
	"schemas": "Schema definitions", "schema": "Schema definitions",
	"tools": "Developer tooling",
}

func scanStructure(repoPath string) []schema.StructureEntry {
	entries, err := os.ReadDir(repoPath)
	if err != nil {
		return nil
	}
	var out []schema.StructureEntry
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") || ignoreDirs[e.Name()] {
			continue
		}
		desc := knownDirs[strings.ToLower(e.Name())]
		if desc == "" {
			desc = "TODO: describe this directory"
		}
		out = append(out, schema.StructureEntry{Path: e.Name() + "/", Description: desc})
	}
	return out
}

// ── Decisions (ADRs) ──────────────────────────────────────────────────────────

var adrDirs = []string{"docs/decisions", "docs/adr", "adr", "doc/architecture/decisions", "docs/rfds", "rfds"}

func scanDecisions(repoPath string, log *[]string) []schema.Decision {
	var decisions []schema.Decision
	for _, dir := range adrDirs {
		entries, err := os.ReadDir(filepath.Join(repoPath, dir))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || strings.HasPrefix(e.Name(), "_") {
				continue
			}
			if d := parseADR(filepath.Join(repoPath, dir, e.Name())); d != nil {
				decisions = append(decisions, *d)
			}
		}
	}
	if len(decisions) > 0 {
		*log = append(*log, "✓  "+itoa(len(decisions))+" decisions from ADRs")
	}
	return decisions
}

func parseADR(path string) *schema.Decision {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(b), "\n")

	title := strings.TrimSuffix(filepath.Base(path), ".md")
	for _, l := range lines {
		if strings.HasPrefix(l, "# ") {
			title = strings.TrimPrefix(l, "# ")
			break
		}
	}

	status := "accepted"
	for _, l := range lines {
		lower := strings.ToLower(l)
		if regexp.MustCompile(`(?i)status\s*[:：]`).MatchString(l) {
			if strings.Contains(lower, "supersed") {
				status = "superseded"
			} else if strings.Contains(lower, "deprecat") {
				status = "deprecated"
			}
			break
		}
	}

	rationale := extractADRSection(lines, []string{"## decision", "## context", "## summary", "## description"})
	if rationale == "" {
		// Fallback: first paragraph after title
		pastTitle := false
		var parts []string
		for _, l := range lines {
			if !pastTitle {
				if strings.HasPrefix(l, "# ") {
					pastTitle = true
				}
				continue
			}
			t := strings.TrimSpace(l)
			if t == "" && len(parts) > 0 {
				break
			}
			if t == "" || strings.HasPrefix(t, "#") {
				continue
			}
			parts = append(parts, t)
		}
		rationale = strings.Join(parts, " ")
	}

	if len(rationale) < 20 {
		return nil
	}
	if len(rationale) > 500 {
		rationale = rationale[:497] + "..."
	}

	d := &schema.Decision{Title: title, Rationale: rationale, Status: status}
	if m := regexp.MustCompile(`(?im)^\*?\*?date\*?\*?\s*[:：]\s*(\d{4}-\d{2}(?:-\d{2})?)`).FindStringSubmatch(string(b)); m != nil {
		d.Date = m[1]
	}
	return d
}

func extractADRSection(lines []string, headers []string) string {
	for i, l := range lines {
		lower := strings.ToLower(l)
		for _, h := range headers {
			if strings.HasPrefix(lower, h) {
				var parts []string
				for _, sl := range lines[i+1:] {
					if strings.HasPrefix(sl, "## ") {
						break
					}
					if t := strings.TrimSpace(sl); t != "" {
						parts = append(parts, t)
					}
				}
				return strings.Join(parts, " ")
			}
		}
	}
	return ""
}

// ── Conventions ───────────────────────────────────────────────────────────────

var reConventionVerb = regexp.MustCompile(`(?i)\b(use|always|never|must|should|ensure|run|write|follow|avoid|prefer|keep|add|include|name|format)\b`)

func scanConventions(repoPath string, log *[]string) []schema.Convention {
	var out []schema.Convention
	for _, name := range []string{"CONTRIBUTING.md", "GOVERNANCE.md"} {
		b, err := os.ReadFile(filepath.Join(repoPath, name))
		if err != nil {
			continue
		}
		out = append(out, extractConventions(string(b), name)...)
	}
	if len(out) > 0 {
		*log = append(*log, "✓  "+itoa(len(out))+" conventions from docs")
	}
	return out
}

func extractConventions(content, source string) []schema.Convention {
	var out []schema.Convention
	reSection := regexp.MustCompile(`(?i)contribut|style|convention|process|standard|rule|format|commit`)
	currentSection := ""
	for _, line := range strings.Split(content, "\n") {
		if m := regexp.MustCompile(`^#{1,3}\s+(.+)`).FindStringSubmatch(line); m != nil {
			currentSection = strings.ToLower(m[1])
			continue
		}
		if !reSection.MatchString(currentSection) {
			continue
		}
		m := regexp.MustCompile(`^\s*[-*]\s+(.+)`).FindStringSubmatch(line)
		if m == nil {
			continue
		}
		text := strings.TrimSpace(m[1])
		if len(text) < 15 || len(text) > 300 {
			continue
		}
		if reConventionVerb.MatchString(text) {
			out = append(out, schema.Convention{Rule: text, Reference: source})
		}
	}
	return out
}

// ── Constraints ───────────────────────────────────────────────────────────────

var reConstraintKw = regexp.MustCompile(`(?i)\b(must|never|always|require|prohibit|shall not|do not|don't|forbidden|mandatory)\b`)
var reConstraintSection = regexp.MustCompile(`(?i)secur|require|rule|constraint|restrict|prohibit`)

func scanConstraints(repoPath string, log *[]string) []schema.Constraint {
	var out []schema.Constraint
	for _, name := range []string{"CONTRIBUTING.md", "SECURITY.md"} {
		b, err := os.ReadFile(filepath.Join(repoPath, name))
		if err != nil {
			continue
		}
		out = append(out, extractConstraints(string(b))...)
	}
	if len(out) > 0 {
		*log = append(*log, "✓  "+itoa(len(out))+" constraints from docs")
	}
	return out
}

func extractConstraints(content string) []schema.Constraint {
	var out []schema.Constraint
	currentSection := ""
	for _, line := range strings.Split(content, "\n") {
		if m := regexp.MustCompile(`^#{1,3}\s+(.+)`).FindStringSubmatch(line); m != nil {
			currentSection = strings.ToLower(m[1])
			continue
		}
		if !reConstraintSection.MatchString(currentSection) {
			continue
		}
		m := regexp.MustCompile(`^\s*[-*]\s+(.+)`).FindStringSubmatch(line)
		if m == nil {
			continue
		}
		text := strings.TrimSpace(m[1])
		if len(text) < 15 || len(text) > 300 {
			continue
		}
		if reConstraintKw.MatchString(text) {
			out = append(out, schema.Constraint{Content: text})
		}
	}
	return out
}

// ── Dependencies ──────────────────────────────────────────────────────────────

func scanDependencies(repoPath string, log *[]string) []schema.Dependency {
	var deps []schema.Dependency
	seen := map[string]bool{}
	add := func(d schema.Dependency) {
		if seen[d.Name] {
			return
		}
		seen[d.Name] = true
		deps = append(deps, d)
	}

	// package.json — org-scoped siblings
	if b, err := os.ReadFile(filepath.Join(repoPath, "package.json")); err == nil {
		var pkg struct {
			Name         string            `json:"name"`
			Dependencies map[string]string `json:"dependencies"`
			DevDeps      map[string]string `json:"devDependencies"`
		}
		if json.Unmarshal(b, &pkg) == nil {
			orgPrefix := ""
			if m := regexp.MustCompile(`^(@[^/]+)/`).FindStringSubmatch(pkg.Name); m != nil {
				orgPrefix = m[1] + "/"
			}
			if orgPrefix != "" {
				all := pkg.Dependencies
				for k, v := range pkg.DevDeps {
					all[k] = v
				}
				for name := range all {
					if strings.HasPrefix(name, orgPrefix) {
						add(schema.Dependency{
							Name:         strings.TrimPrefix(name, orgPrefix),
							URL:          "https://www.npmjs.com/package/" + name,
							Relationship: "depends_on",
							Why:          "npm package " + name,
						})
					}
				}
			}
		}
	}

	// go.mod — same-org siblings
	if b, err := os.ReadFile(filepath.Join(repoPath, "go.mod")); err == nil {
		content := string(b)
		if m := regexp.MustCompile(`(?m)^module\s+(\S+)`).FindStringSubmatch(content); m != nil {
			parts := strings.SplitN(m[1], "/", 3)
			n := 2
			if len(parts) < n {
				n = len(parts)
			}
			orgBase := strings.Join(parts[:n], "/")
			for _, match := range regexp.MustCompile(`(?m)^\s+(\S+)\s`).FindAllStringSubmatch(content, -1) {
				dep := match[1]
				if strings.HasPrefix(dep, orgBase+"/") && dep != m[1] {
					name := dep[strings.LastIndex(dep, "/")+1:]
					add(schema.Dependency{
						Name:         name,
						URL:          "https://" + dep,
						Relationship: "depends_on",
					})
				}
			}
		}
	}

	// docker-compose depends_on
	for _, name := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"} {
		b, err := os.ReadFile(filepath.Join(repoPath, name))
		if err != nil {
			continue
		}
		for _, svc := range composeDepends(b) {
			add(schema.Dependency{
				Name:         svc,
				URL:          "https://example.com/REPLACE-ME/" + svc,
				Relationship: "calls",
				Why:          "docker-compose service dependency",
			})
		}
	}

	if len(deps) > 0 {
		*log = append(*log, "✓  "+itoa(len(deps))+" dependencies from manifest")
	}
	return deps
}

func composeDepends(src []byte) []string {
	var doc struct {
		Services map[string]struct {
			DependsOn any `yaml:"depends_on"`
		} `yaml:"services"`
	}
	if err := yaml.Unmarshal(src, &doc); err != nil {
		return nil
	}
	seen := map[string]bool{}
	for _, svc := range doc.Services {
		switch v := svc.DependsOn.(type) {
		case []any:
			for _, name := range v {
				if s, ok := name.(string); ok {
					seen[s] = true
				}
			}
		case map[string]any:
			for name := range v {
				seen[name] = true
			}
		}
	}
	var out []string
	for name := range seen {
		out = append(out, name)
	}
	return out
}

// ── Links ─────────────────────────────────────────────────────────────────────

func scanLinks(repoPath string, id schema.Identity) schema.Links {
	links := make(schema.Links)
	if id.URL == "" || strings.Contains(id.URL, "REPLACE-ME") {
		return links
	}
	links["repository"] = id.URL
	if strings.Contains(id.URL, "github.com") {
		links["issues"] = id.URL + "/issues"
		if _, err := os.Stat(filepath.Join(repoPath, ".github", "workflows")); err == nil {
			links["ci"] = id.URL + "/actions"
		}
	} else if strings.Contains(id.URL, "gitlab") {
		links["issues"] = id.URL + "/-/issues"
		if _, err := os.Stat(filepath.Join(repoPath, ".gitlab-ci.yml")); err == nil {
			links["ci"] = id.URL + "/-/pipelines"
		}
	}
	return links
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func gitCmd(repoPath string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", repoPath}, args...)...).Output()
	return string(out), err
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
