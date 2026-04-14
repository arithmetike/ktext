# Contributing to ktext

## Philosophy

ktext is a single-purpose tool: it reads and writes files. No database, no server, no network calls, no LLM. Changes that add runtime dependencies or scope creep toward a service model will be declined regardless of implementation quality.

Keep it small. Keep it fast. Keep it honest.

## Getting started

```bash
git clone https://github.com/arithmetike/ktext
cd ktext
go build ./cmd/ktext         # build
go test ./...                 # test
go vet ./...                  # vet
```

Go 1.25+ required. No other tooling needed.

## Project layout

```
cmd/ktext/          CLI entry point and subcommand handlers
internal/schema/    Types, JSON Schema validator, quality rules, scorer
internal/export/    Format renderers (yaml, xml, json)
internal/scan/      Filesystem scanner for ktext init
```

Everything in `internal/` is unexported by design. The only public surface is the binary.

## Making changes

### Schema changes

`internal/schema/context-yaml.schema.json` is the canonical source of truth. If you add or rename a field:

1. Update the JSON Schema first.
2. Update the Go types in `context.go` to match exactly.
3. Update the scorer in `score.go` if the new field should affect quality scoring.
4. Update `CONTEXT.yaml` in this repo if the field is applicable.
5. Never rename an existing field without a documented migration path — existing files break silently.

### Adding an export format

Export formats live in `internal/export/`. To add one:

1. Create `internal/export/<name>.go` with a `render<Name>(doc *schema.Context) string` function.
2. Add an entry to `All` in `render.go` and a case to the `Render()` switch.
3. Do not add tool-specific formats (claude-md, cursorrules, copilot-instructions, etc.) — the job of ktext is to export structured data, not to render tool configuration.

### Scanner heuristics

`internal/scan/scan.go` discovers project context from the filesystem. Scanner output is written immediately by default; use `-interactive` to review before writing. False positives are acceptable. False negatives (missing real information) are worse. When in doubt, surface the data and let the user decide.

## Tests

Tests live alongside the code they test, as external test packages (`package schema_test`, not `package schema`). This keeps tests honest about the public API.

```bash
go test ./...                          # all tests
go test -run TestParse ./internal/schema/   # one test
go test -v ./...                       # verbose
```

There are no mocks. Tests use real YAML/JSON inputs.

## Flags and CLI conventions

- Each subcommand gets its own `flag.NewFlagSet`.
- All commands route through `splitArgs()` before parsing — this handles the stdlib `flag` limitation of stopping at the first non-flag argument.
- Commands return `int` exit codes. Only `main()` calls `os.Exit`. This makes command functions directly testable.
- Flags come before positional args in documentation, even though `splitArgs()` allows either order.

## Pull requests

- One logical change per PR. Refactors and features in separate PRs.
- `go test ./...` and `go vet ./...` must pass.
- Run `ktext validate` on the repo and confirm the score does not drop.
- Update `CONTEXT.yaml` if your change affects constraints, decisions, conventions, risks, or structure.

## What will not be merged

- Runtime dependencies beyond the current three (`yaml.v3`, `jsonschema`, `x/text`).
- Network calls, database access, or LLM integration of any kind.
- Tool-specific export formats (IDE configs, agent instructions, etc.).
- Features that require a running process or daemon.
- Breaking schema changes without a migration path.

## License

MIT. By contributing you agree your changes will be licensed under the same terms.
