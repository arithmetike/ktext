# ktext

Every project has context that lives only in the heads of its engineers — why Postgres instead of MySQL, what you must never log, which directory owns what. When a new engineer joins, or an AI agent opens a PR, that context is missing. They guess. They get it wrong.

`CONTEXT.yaml` is a single file that captures that knowledge in a structured, machine-readable format. `ktext` generates it, scores it, and exports it.

```
$ ktext init
Scanning repo...
  ✓  Identity: my-service (service)
  ✓  Purpose from README
  ✓  Build commands from package.json (3 scripts)
  ✓  Directory structure (6 directories)
  ✓  4 decisions from ADRs

Review discovered values (Enter to accept, or type a new value):
  name       my-service
  url        https://github.com/org/my-service
  purpose    ...
  ...

Written to CONTEXT.yaml (score: 72/100)
Run 'ktext validate' for detailed feedback.
```

## Install

```bash
go install github.com/arithmetike/ktext/cmd/ktext@latest
```

Or download a binary from the [releases page](https://github.com/arithmetike/ktext/releases).

## Commands

### `ktext init`

Scans the repository and generates a `CONTEXT.yaml` file. The scanner reads your README, package manifests, Makefile, ADRs, directory structure, and CONTRIBUTING.md to make initial guesses, then walks you through each field interactively so you can confirm or correct them. Hit Enter to accept a discovered value, or type a replacement.

If a `CONTEXT.yaml` already exists, `ktext init` will not overwrite it. Delete it first or edit it directly.

```bash
ktext init
```

### `ktext validate`

Parses and scores the `CONTEXT.yaml` in the current directory. Each of the eight sections (identity, constraints, decisions, conventions, risks, dependencies, working, ownership) is scored independently. The output shows what passed, what needs work, and specific fixes you can make.

Exits `0` if the score meets the threshold, `1` if it does not. The default threshold is 80. Use this in CI to enforce a minimum quality bar.

```bash
ktext validate                  # score against default threshold (80)
ktext validate -threshold 90    # stricter threshold
ktext validate -json            # machine-readable output
```

### `ktext export`

Renders `CONTEXT.yaml` to a different format. By default, output goes to stdout so you can pipe it or inspect it before writing. Use `-write` to write directly to the conventional output file for that format.

```bash
ktext export xml                # render to stdout
ktext export -write xml         # write to context.xml
ktext export json               # render to stdout
ktext export -list              # list all supported formats
```

## Export formats

| Format | Output file | Use case |
|---|---|---|
| `yaml` | `CONTEXT.yaml` | Re-serialize and normalize the source file |
| `xml` | `context.xml` | Compact token-efficient format for LLM context injection |
| `json` | `context.json` | Structured data for tooling and scripts |

## CONTEXT.yaml

The file is designed to be read by both humans and machines. Each section serves a specific purpose:

- **identity** — what the project is, what it does, and where it lives
- **constraints** — hard rules that must never be violated (compliance, architecture, safety)
- **decisions** — why the project is built the way it is; the rationale behind major choices
- **conventions** — coding and process rules that apply throughout the codebase
- **risks** — known fragile areas, tech debt, and operational concerns with mitigations
- **dependencies** — external systems and libraries this project relies on
- **working** — build commands, directory layout, and dev environment notes
- **ownership** — who maintains it and how to escalate

```yaml
ktext: "1.0.0"
identity:
  name: my-service
  url: https://github.com/org/my-service
  type: service
  status: active
  purpose: Handles payment processing for the checkout flow.

constraints:
  - content: Never log raw card data or PII in any request handler
    why: PCI compliance

decisions:
  - title: Use PostgreSQL for all persistent state
    rationale: Need transactional integrity and JSONB support
    status: accepted

conventions:
  - rule: All DB queries go through the repository layer — no raw SQL in handlers

risks:
  - content: Migration rollback untested beyond 2 versions
    severity: high
    mitigation: Add rollback smoke test to CI before next release

dependencies:
  - name: PostgreSQL
    url: https://www.postgresql.org
    why: Primary data store

working:
  commands:
    - command: make build
      description: build the service
    - command: make test
      description: run tests
  structure:
    - path: internal/payments/
      description: payment processing logic and provider integrations
  notes:
    - Run migrations before starting the server locally
```

Full schema reference: [context-yaml.schema.json](internal/schema/context-yaml.schema.json)

## Use in CI

```yaml
# .github/workflows/context.yml
- name: Validate CONTEXT.yaml
  run: |
    go install github.com/arithmetike/ktext/cmd/ktext@latest
    ktext validate -threshold 80
```

A score below the threshold exits `1` and fails the step. The output tells you exactly which sections need work and what to fix.

## No database. No server. No account.

`ktext` reads and writes files. The schema and scoring logic are embedded in the binary. There is nothing to sign up for, nothing running in the background, and nothing phoning home.

## License

MIT
