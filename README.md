# ktext

Every codebase has context that lives in engineers' heads: why Postgres instead of MySQL, what you must never log, which layer owns what. When a new engineer joins, or an AI agent opens a PR, that context is missing. They guess. They get it wrong.

`CONTEXT.yaml` is a single file that captures that knowledge in a structured, machine-readable format. `ktext` generates it, validates it, and exports it.

**[Full documentation at ktext.dev](https://ktext.dev)**

## Install

```bash
go install github.com/arithmetike/ktext/cmd/ktext@latest
```

Or download a binary from the [releases page](https://github.com/arithmetike/ktext/releases).

## Commands

```bash
ktext init              # scan the repo and generate CONTEXT.yaml
ktext validate          # score it out of 100
ktext validate -json    # machine-readable output for CI
ktext export xml        # export as compact XML for LLM injection
ktext export json       # export as JSON
```

## Use in CI

```yaml
- name: Validate CONTEXT.yaml
  run: |
    go install github.com/arithmetike/ktext/cmd/ktext@latest
    ktext validate -threshold 80
```

## License

MIT — no accounts, no strings attached.
