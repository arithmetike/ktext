# ktext

Every codebase has context that lives in engineers' heads: why Postgres instead of MySQL, what you must never log, which layer owns what. Without it, agents burn thousands of tokens exploring the codebase to reconstruct what you already know. You end up steering them manually, repeating yourself in every session, or writing lengthy prompt instructions that go stale.

`CONTEXT.yaml` is a single file that captures that knowledge in a structured, machine-readable format. `ktext` generates it, validates it, and scores it. Load it once and your agent starts every session already up to speed on your codebase.

**[Full documentation at ktext.dev](https://ktext.dev)**

## Install

```bash
go install github.com/arithmetike/ktext/cmd/ktext@latest
```

Or download a binary from the [releases page](https://github.com/arithmetike/ktext/releases).

## Commands

### `ktext init`

Scans the repo and generates a `CONTEXT.yaml`. Reads your README, package manifests, Makefile, ADRs, and directory structure to fill in what it can, writes the file immediately, and reports what was discovered and what still needs manual editing. Use `-interactive` to review and edit each value before writing.

```bash
ktext init                  # scan and write immediately
ktext init -interactive     # review each value before writing
```

### `ktext validate`

Scores your `CONTEXT.yaml` out of 100 across eight sections: identity, constraints, decisions, conventions, risks, dependencies, working, and ownership. Each section is evaluated for presence, completeness, and quality of language. The output shows what passed, what failed, and exactly what to fix.

Exits `0` if the score meets the threshold, `1` if not. Use it as a CI gate.

```bash
ktext validate                  # score against the default threshold (80)
ktext validate -threshold 90    # stricter bar
ktext validate -json            # machine-readable output
```

### `ktext export`

Renders `CONTEXT.yaml` to a different format. The XML output is compact and optimized for injection into an LLM context window, roughly 80% fewer tokens than having an agent explore the codebase manually.

```bash
ktext export xml        # compact XML for LLM injection
ktext export json       # structured JSON for tooling
ktext export -list      # list all supported formats
```

## The schema

`CONTEXT.yaml` is validated against a [JSON Schema (draft 2020-12)](internal/schema/context-yaml.schema.json) embedded in the binary. The eight sections capture what matters most for an agent working in your codebase:

| Section | Purpose |
|---|---|
| `identity` | What the project is and what it does |
| `constraints` | Hard rules that must never be violated |
| `decisions` | Architectural choices and their rationale |
| `conventions` | Coding and process rules |
| `risks` | Known fragile areas and mitigations |
| `dependencies` | External systems and why they're used |
| `working` | Build commands and directory layout |
| `ownership` | Who maintains it and how to escalate |

Full schema reference: [ktext.dev/docs/schema](https://ktext.dev/docs/schema)

## Load context automatically in Claude Code

Add a `SessionStart` hook to `.claude/settings.json` at the root of your project. Claude Code runs it at the start of every session and injects the output into context automatically.

```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup|compact",
        "hooks": [
          {
            "type": "command",
            "command": "ktext export xml"
          }
        ]
      }
    ]
  }
}
```

For other tools (Cursor, Copilot, Windsurf), see [ktext.dev/docs](https://ktext.dev/docs).

## Use in CI

```yaml
- name: Validate CONTEXT.yaml
  run: |
    go install github.com/arithmetike/ktext/cmd/ktext@latest
    ktext validate -threshold 80
```

## License

MIT. No accounts, no strings attached.
