# ktext

Go CLI tool — single binary, no server, no database, no LLM dependency.

## Quick reference

```bash
go build ./cmd/ktext          # build binary
go test ./...                 # run all tests
go vet ./...                  # vet sources
ktext validate                # score this repo's own CONTEXT.yaml
ktext export xml              # render context for LLM injection
```

**Always build before manual testing.** The globally installed binary is not rebuilt automatically.

## Architecture

- `cmd/ktext/` — CLI entry point, subcommand handlers (init.go, validate.go, export.go), flag parser (args.go)
- `internal/schema/` — Go types (context.go), JSON Schema validator (parse.go), quality checker (quality.go), scorer (score.go)
- `internal/export/` — format dispatch (render.go), XML/YAML/JSON renderers
- `internal/scan/` — filesystem scanner used by `ktext init`

## Critical patterns

**splitArgs()** — Go stdlib `flag.Parse` stops at the first non-flag argument. Every command that takes both flags and positional args must route through `splitArgs()` in args.go first, or put flags before positional args.

**go:embed for the schema** — the JSON Schema is embedded in the binary at compile time. Never load it from disk at runtime. Never construct a path to it.

**Schema is canonical** — `internal/schema/context-yaml.schema.json` is the source of truth. Go types in `context.go` mirror it but do not define it. If they diverge, the schema wins. Always update both together.

**External test packages** — tests use `package foo_test` (not `package foo`). This tests the exported API, not internals.

**Exit codes** — command functions return `int` exit codes. Only `main()` calls `os.Exit`. This keeps commands testable without spawning subprocesses.

**min(score, max)** — always cap section scores before returning a `SectionResult`. Floating-point rounding can push scores over the ceiling.

## Flags use single dash

Go stdlib `flag` package uses single-dash convention (`-threshold`, `-json`). This is consistent with Go tooling (`go test -run`, `go build -o`). Both `-flag` and `--flag` work at runtime. Do not switch to cobra or urfave/cli — the three subcommands don't justify the dependency.

## What not to add

- No database, server, or LLM dependency — ever
- No tool-specific export formats (claude-md, cursorrules, copilot) — tool-specific rendering belongs in consuming tools
- No runtime schema loading from disk — go:embed only

## Doc site

Any change to CLI commands, flags, schema fields, scoring logic, or export formats must be reflected on [ktext.dev](https://ktext.dev) (repo: `ktext-dev`) before or alongside the release.

## Releases

GoReleaser produces binaries for linux/darwin/windows × amd64/arm64. Triggered by pushing a `v*` tag. CI runs on every push and validates this repo's own CONTEXT.yaml at threshold 90.
