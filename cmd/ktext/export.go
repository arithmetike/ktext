package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arithmetike/ktext/internal/export"
	"github.com/arithmetike/ktext/internal/schema"
)

func runExport(args []string) int {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	file := fs.String("file", "CONTEXT.yaml", "Path to the CONTEXT.yaml file")
	write := fs.Bool("write", false, "Write output to the format's conventional filename instead of stdout")
	list := fs.Bool("list", false, "List available formats and exit")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: ktext export [format] [flags]

Render CONTEXT.yaml to a target format and print to stdout.
Use -write to write to the format's conventional filename instead.

Formats:
`)
		nameW := 0
		for _, f := range export.All {
			if len(f.Name) > nameW {
				nameW = len(f.Name)
			}
		}
		for _, f := range export.All {
			fmt.Fprintf(os.Stderr, "  %-*s  %s\n", nameW, f.Name, f.Filename)
		}
		fmt.Fprint(os.Stderr, "\nFlags:\n")
		fs.PrintDefaults()
	}
	flagArgs, positional := splitArgs(args, map[string]bool{"write": true, "list": true})
	if err := fs.Parse(flagArgs); err != nil {
		return 1
	}

	if *list {
		nameW := 0
		for _, f := range export.All {
			if len(f.Name) > nameW {
				nameW = len(f.Name)
			}
		}
		for _, f := range export.All {
			fmt.Printf("%-*s  %s\n", nameW, f.Name, f.Filename)
		}
		return 0
	}

	if len(positional) == 0 {
		fmt.Fprintln(os.Stderr, "ktext export: format argument required")
		fmt.Fprintln(os.Stderr)
		fs.Usage()
		return 1
	}

	format := strings.ToLower(positional[0])
	var matched *export.Format
	for i, f := range export.All {
		if f.Name == format {
			matched = &export.All[i]
			break
		}
	}
	if matched == nil {
		fmt.Fprintf(os.Stderr, "ktext export: unknown format %q\n\nRun 'ktext export -list' to see available formats.\n", format)
		return 1
	}

	src, err := os.ReadFile(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ktext export: cannot read %s: %v\n", *file, err)
		return 1
	}

	result := schema.Parse(src)
	if !result.OK() {
		fmt.Fprintln(os.Stderr, "Invalid CONTEXT.yaml:")
		for _, e := range result.Errors {
			if len(e.Path) > 0 {
				fmt.Fprintf(os.Stderr, "  %s: %s\n", strings.Join(e.Path, "."), e.Message)
			} else {
				fmt.Fprintf(os.Stderr, "  %s\n", e.Message)
			}
		}
		return 1
	}

	out, err := export.Render(format, result.Doc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ktext export: render error: %v\n", err)
		return 1
	}

	if *write {
		dest := matched.Filename
		if dir := filepath.Dir(dest); dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "ktext export: cannot create directory %s: %v\n", dir, err)
				return 1
			}
		}
		if err := os.WriteFile(dest, []byte(out), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "ktext export: cannot write %s: %v\n", dest, err)
			return 1
		}
		fmt.Printf("Written to %s\n", dest)
		return 0
	}

	fmt.Print(out)
	return 0
}
