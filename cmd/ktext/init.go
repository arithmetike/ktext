package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/arithmetike/ktext/internal/scan"
	"github.com/arithmetike/ktext/internal/schema"
	"gopkg.in/yaml.v3"
)

func runInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	file := fs.String("file", "CONTEXT.yaml", "Path to write the generated file")
	interactive := fs.Bool("interactive", false, "Review and edit discovered values before writing")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: ktext init [flags]

Scan the current directory, generate a CONTEXT.yaml, and report what was
discovered and what still needs to be filled in manually.

Flags:
`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 1
	}

	repoPath, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ktext init: %v\n", err)
		return 1
	}

	// Confirm overwrite if file exists (always prompt — not behind a flag)
	if _, err := os.Stat(*file); err == nil {
		if !promptYesNo(fmt.Sprintf("%s already exists. Overwrite?", *file)) {
			fmt.Println("Aborted.")
			return 0
		}
	}

	fmt.Println("\nScanning repo...")
	result := scan.Repo(repoPath)
	for _, line := range result.Log {
		fmt.Println("  " + line)
	}
	fmt.Println()

	doc := buildDoc(result)

	if *interactive {
		if err := interactiveReview(&doc); err != nil {
			fmt.Fprintf(os.Stderr, "ktext init: %v\n", err)
			return 1
		}
	}

	out, err := serializeDoc(doc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ktext init: serialize error: %v\n", err)
		return 1
	}

	if err := os.WriteFile(*file, []byte(out), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "ktext init: cannot write %s: %v\n", *file, err)
		return 1
	}

	score := schema.ScoreDocument(&doc)
	fmt.Printf("Written to %s — %d/100\n\n", *file, score.Score)
	printSections(doc)
	return 0
}

// ── Document builder ─────────────────────────────────────────────────────────

func buildDoc(r scan.Result) schema.Context {
	doc := schema.Context{
		Ktext:    "1.0.0",
		Identity: r.Identity,
	}
	if len(r.Constraints) > 0 {
		doc.Constraints = r.Constraints
	}
	if len(r.Decisions) > 0 {
		doc.Decisions = r.Decisions
	}
	if len(r.Conventions) > 0 {
		doc.Conventions = r.Conventions
	}
	if len(r.Dependencies) > 0 {
		doc.Dependencies = r.Dependencies
	}
	if len(r.Working.Commands) > 0 || len(r.Working.Structure) > 0 || len(r.Working.Notes) > 0 {
		w := r.Working
		doc.Working = &w
	}
	if len(r.Links) > 0 {
		doc.Links = r.Links
	}
	return doc
}

// ── Serialization ─────────────────────────────────────────────────────────────

func serializeDoc(doc schema.Context) (string, error) {
	b, err := yaml.Marshal(doc)
	if err != nil {
		return "", err
	}
	header := "# CONTEXT.yaml — machine-readable context for this repository\n" +
		"# Schema: https://contextfile.org/schema/v1\n\n"
	return header + string(b), nil
}

// ── Section report ────────────────────────────────────────────────────────────

type sectionStatus struct {
	label  string
	filled bool
	detail string
	hint   string
}

func printSections(doc schema.Context) {
	sections := []sectionStatus{
		{
			label:  "identity",
			filled: doc.Identity.Name != "",
			detail: fmt.Sprintf("%s (%s · %s)", doc.Identity.Name, doc.Identity.Type, doc.Identity.Status),
		},
		{
			label:  "purpose",
			filled: doc.Identity.Purpose != "",
			detail: truncate(doc.Identity.Purpose, 72),
			hint:   "One sentence: what does this project do and who uses it?",
		},
		{
			label:  "constraints",
			filled: len(doc.Constraints) > 0,
			detail: fmt.Sprintf("%d found", len(doc.Constraints)),
			hint:   "Rules that must never be violated — architecture guardrails, security rules, \"never do X\".",
		},
		{
			label:  "decisions",
			filled: len(doc.Decisions) > 0,
			detail: fmt.Sprintf("%d found", len(doc.Decisions)),
			hint:   "Why key choices were made — the rationale behind architectural decisions, not just what they are.",
		},
		{
			label:  "conventions",
			filled: len(doc.Conventions) > 0,
			detail: fmt.Sprintf("%d found", len(doc.Conventions)),
			hint:   "How code is written here — naming, patterns, style rules specific to this repo.",
		},
		{
			label:  "risks",
			filled: len(doc.Risks) > 0,
			detail: fmt.Sprintf("%d found", len(doc.Risks)),
			hint:   "Known fragile areas, operational gotchas, or dangerous patterns to be aware of.",
		},
		{
			label:  "dependencies",
			filled: len(doc.Dependencies) > 0,
			detail: fmt.Sprintf("%d found", len(doc.Dependencies)),
			hint:   "Key libraries and why they were chosen — not everything, just the ones that shape the code.",
		},
		{
			label:  "commands",
			filled: doc.Working != nil && len(doc.Working.Commands) > 0,
			detail: func() string {
				if doc.Working != nil {
					return fmt.Sprintf("%d found", len(doc.Working.Commands))
				}
				return ""
			}(),
			hint: "Build, test, lint, and dev commands for this repo.",
		},
		{
			label:  "structure",
			filled: doc.Working != nil && len(doc.Working.Structure) > 0,
			detail: func() string {
				if doc.Working != nil {
					return fmt.Sprintf("%d entries", len(doc.Working.Structure))
				}
				return ""
			}(),
			hint: "Key directories and what lives in each one.",
		},
		{
			label:  "ownership",
			filled: doc.Ownership != nil,
			hint:   "Who maintains this, and where to escalate — team name, Slack channel, or on-call link.",
		},
	}

	var filled, gaps []sectionStatus
	for _, s := range sections {
		if s.filled {
			filled = append(filled, s)
		} else {
			gaps = append(gaps, s)
		}
	}

	if len(filled) > 0 {
		fmt.Println("Discovered:")
		for _, s := range filled {
			if s.detail != "" {
				fmt.Printf("  ✓  %-14s %s\n", s.label, s.detail)
			} else {
				fmt.Printf("  ✓  %s\n", s.label)
			}
		}
	}

	if len(gaps) > 0 {
		fmt.Println("\nFill in manually to improve your score:")
		for _, s := range gaps {
			fmt.Printf("  ✗  %-14s %s\n", s.label, s.hint)
		}
	}

	fmt.Println("\nRun 'ktext validate -verbose' for per-section scoring.")
	if len(gaps) > 0 {
		fmt.Println("Use -interactive to review and edit discovered values before writing.")
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

// ── Interactive review ────────────────────────────────────────────────────────

func interactiveReview(doc *schema.Context) error {
	r := bufio.NewReader(os.Stdin)
	ask := func(prompt string) (string, error) {
		fmt.Print(prompt)
		line, err := r.ReadString('\n')
		return strings.TrimRight(line, "\r\n"), err
	}

	// field shows the current value and prompts for a replacement.
	// Returns current if the user presses Enter without typing.
	field := func(label, current, hint string) (string, error) {
		display := current
		if display == "" {
			display = "\033[90m(not found)\033[0m"
		}
		hintStr := ""
		if hint != "" {
			hintStr = "  \033[90m" + hint + "\033[0m"
		}
		fmt.Printf("  %-10s %s%s\n             → ", label, display, hintStr)
		input, err := r.ReadString('\n')
		if err != nil {
			return current, err
		}
		if v := strings.TrimRight(input, "\r\n"); v != "" {
			return v, nil
		}
		return current, nil
	}

	fmt.Println("Review discovered values (Enter to accept, or type a new value):")
	fmt.Println()

	var err error
	doc.Identity.Name, err = field("name", doc.Identity.Name, "")
	if err != nil {
		return err
	}

	urlDisplay := doc.Identity.URL
	if strings.Contains(urlDisplay, "REPLACE-ME") {
		urlDisplay = ""
	}
	doc.Identity.URL, err = field("url", urlDisplay, "(canonical repo URL)")
	if err != nil {
		return err
	}
	if doc.Identity.URL == "" {
		doc.Identity.URL = "https://example.com/REPLACE-ME"
	}

	doc.Identity.Purpose, err = field("purpose", doc.Identity.Purpose, "(what does this project do?)")
	if err != nil {
		return err
	}

	typeVal, err := field("type", doc.Identity.Type, "(service/library/application/tool/component/platform/documentation/ci/other)")
	if err != nil {
		return err
	}
	validTypes := map[string]bool{"service": true, "library": true, "application": true, "component": true, "platform": true, "tool": true, "documentation": true, "ci": true, "other": true}
	if validTypes[typeVal] {
		doc.Identity.Type = typeVal
	}

	statusVal, err := field("status", doc.Identity.Status, "(active/deprecated/experimental/archived)")
	if err != nil {
		return err
	}
	validStatuses := map[string]bool{"active": true, "deprecated": true, "experimental": true, "archived": true}
	if validStatuses[statusVal] {
		doc.Identity.Status = statusVal
	}

	// Offer to fill missing high-value sections
	var gaps []string
	if len(doc.Constraints) == 0 {
		gaps = append(gaps, "constraints")
	}
	if len(doc.Risks) == 0 {
		gaps = append(gaps, "risks")
	}
	if doc.Ownership == nil {
		gaps = append(gaps, "ownership")
	}

	if len(gaps) > 0 {
		fmt.Printf("\nGaps: %s\n", strings.Join(gaps, ", "))
		doFill, err := ask("Fill missing sections now? [y/N] ")
		if err != nil {
			return err
		}
		if strings.ToLower(doFill) == "y" {
			if err := fillGaps(doc, gaps, ask); err != nil {
				return err
			}
		}
	}

	score := schema.ScoreDocument(doc)
	fmt.Printf("\nScore: %d/100\n", score.Score)
	return nil
}

func fillGaps(doc *schema.Context, gaps []string, ask func(string) (string, error)) error {
	for _, gap := range gaps {
		switch gap {
		case "constraints":
			fmt.Println("\n# Constraints — what must an engineer or AI never do in this codebase?")
			for {
				input, err := ask("Constraint (Enter to skip): ")
				if err != nil || input == "" {
					break
				}
				doc.Constraints = append(doc.Constraints, schema.Constraint{Content: input})
				fmt.Println("  ✓  Added")
			}

		case "risks":
			fmt.Println("\n# Risks — what's fragile, broken, or dangerous here?")
			for {
				input, err := ask("Risk (Enter to skip): ")
				if err != nil || input == "" {
					break
				}
				sev, err := ask("  Severity [high/medium/low] (default: medium): ")
				if err != nil {
					break
				}
				if sev == "" || (sev != "high" && sev != "medium" && sev != "low") {
					sev = "medium"
				}
				doc.Risks = append(doc.Risks, schema.Risk{Content: input, Severity: sev})
				fmt.Println("  ✓  Added")
			}

		case "ownership":
			fmt.Println()
			team, err := ask("Owning team (Enter to skip): ")
			if err != nil || team == "" {
				continue
			}
			doc.Ownership = &schema.Ownership{Team: team}
			escalation, err := ask("Escalation contact (Slack, email, PagerDuty): ")
			if err == nil && escalation != "" {
				doc.Ownership.Escalation = escalation
			}
			fmt.Println("  ✓  Added")
		}
	}
	return nil
}

// ── Utilities ─────────────────────────────────────────────────────────────────

func promptYesNo(question string) bool {
	fmt.Printf("%s [y/N] ", question)
	r := bufio.NewReader(os.Stdin)
	input, _ := r.ReadString('\n')
	answer := strings.ToLower(strings.TrimRight(input, "\r\n"))
	return answer == "y" || answer == "yes"
}

