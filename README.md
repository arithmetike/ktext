# ktext

Generate, validate, and export `CONTEXT.yaml` — a machine-readable project context file that AI coding agents and developer tools can read.

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

| Command | What it does |
|---|---|
| `ktext init` | Scan the repo and generate `CONTEXT.yaml` |
| `ktext validate` | Score quality; exits 1 if score < threshold |
| `ktext validate -threshold 80` | Set minimum passing score |
| `ktext export xml` | Render to XML |
| `ktext export -list` | List all supported formats |
| `ktext export -write json` | Write to the conventional filename |

## Export formats

| Format | Output file |
|---|---|
| `yaml` | `CONTEXT.yaml` |
| `xml` | `context.xml` |
| `json` | `context.json` |

## CONTEXT.yaml

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
    relationship: depends_on
    why: Primary data store

working:
  commands:
    - command: make build
      description: build the service
    - command: make test
      description: run tests
```

Full schema reference: [schema/context-yaml.schema.json](internal/schema/context-yaml.schema.json)

## No database. No server. No account.

`ktext` reads and writes files. That's it. The schema and scoring logic are embedded in the binary.

## License

MIT
