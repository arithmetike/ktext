package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/arithmetike/ktext/internal/schema"
)

func runValidate(args []string) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	file := fs.String("file", "CONTEXT.yaml", "Path to the CONTEXT.yaml file")
	threshold := fs.Int("threshold", 80, "Minimum passing score (0–100)")
	verbose := fs.Bool("verbose", false, "Show per-section detail")
	asJSON := fs.Bool("json", false, "Output results as JSON")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: ktext validate [flags]

Score CONTEXT.yaml quality and suggest improvements.
Exits 0 if score >= threshold, 1 otherwise.

Flags:
`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 1
	}

	src, err := os.ReadFile(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ktext validate: cannot read %s: %v\n", *file, err)
		return 1
	}

	result := schema.Parse(src)
	if !result.OK() {
		fmt.Fprintf(os.Stderr, "Invalid CONTEXT.yaml:\n")
		for _, e := range result.Errors {
			if len(e.Path) > 0 {
				fmt.Fprintf(os.Stderr, "  %s: %s\n", formatPath(e.Path), e.Message)
			} else {
				fmt.Fprintf(os.Stderr, "  %s\n", e.Message)
			}
		}
		return 1
	}

	score := schema.ScoreDocument(result.Doc)
	pass := score.Score >= *threshold

	if *asJSON {
		return printJSON(score, *threshold, pass)
	}

	printText(score, *file, *threshold, *verbose, pass)
	return boolToInt(!pass)
}

// ── Output ───────────────────────────────────────────────────────────────────

func printText(score schema.ScoreResult, file string, threshold int, verbose, pass bool) {
	fmt.Printf("%s — %d/100\n\n", file, score.Score)

	// Column widths
	nameW := 0
	for _, s := range score.Sections {
		if len(s.Name) > nameW {
			nameW = len(s.Name)
		}
	}
	maxW := len("10") // widest max value
	_ = maxW

	for _, s := range score.Sections {
		icon := statusIcon(s.Status)
		fmt.Printf("  %s  %-*s  %2d/%-2d\n", icon, nameW, s.Name, s.Score, s.Max)
		if verbose {
			for _, d := range s.Details {
				fmt.Printf("       %-*s    %s\n", nameW, "", d)
			}
		}
	}

	if len(score.Fixes) > 0 {
		fmt.Println("\nSuggested fixes:")
		for _, f := range score.Fixes {
			fmt.Printf("  • %s\n", f)
		}
	}

	fmt.Println()
	if pass {
		fmt.Printf("PASS  %d/%d\n", score.Score, threshold)
	} else {
		fmt.Printf("FAIL  score %d is below threshold %d\n", score.Score, threshold)
	}
}

func printJSON(score schema.ScoreResult, threshold int, pass bool) int {
	out := struct {
		Score     int                    `json:"score"`
		Threshold int                    `json:"threshold"`
		Pass      bool                   `json:"pass"`
		Sections  []schema.SectionResult `json:"sections"`
		Fixes     []string               `json:"fixes"`
	}{
		Score:     score.Score,
		Threshold: threshold,
		Pass:      pass,
		Sections:  score.Sections,
		Fixes:     score.Fixes,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "ktext validate: JSON encode error: %v\n", err)
		return 1
	}
	if !pass {
		return 1
	}
	return 0
}

func statusIcon(status string) string {
	switch status {
	case "pass":
		return "✓"
	case "warn":
		return "⚠"
	default:
		return "✗"
	}
}

func formatPath(segments []string) string {
	return strings.Join(segments, ".")
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
