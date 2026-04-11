package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		usage(os.Stdout)
		os.Exit(1)
	}

	cmd, args := os.Args[1], os.Args[2:]
	switch cmd {
	case "init":
		os.Exit(runInit(args))
	case "validate":
		os.Exit(runValidate(args))
	case "export":
		os.Exit(runExport(args))
	case "version", "--version", "-v":
		fmt.Println("ktext " + version)
	case "help", "--help", "-h":
		usage(os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "ktext: unknown command %q\n\nRun 'ktext help' for usage.\n", cmd)
		os.Exit(1)
	}
}
