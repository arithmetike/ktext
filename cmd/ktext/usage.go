package main

import (
	"fmt"
	"io"
)

func usage(w io.Writer) {
	fmt.Fprint(w, `ktext — structured context for your codebase

Usage:
  ktext <command> [flags]

Commands:
  init        Generate CONTEXT.yaml from the current directory
  validate    Score CONTEXT.yaml quality and suggest improvements
  export      Render CONTEXT.yaml to a target format

Run 'ktext <command> -help' for command-specific flags.
`)
}
