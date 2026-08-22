# Testing

How to verify dbt-guard behavior locally and in CI pipelines.

---

## Unit tests

From repository root:

```bash
go test ./...
```

Coverage:

| Package | Scope |
|---------|--------|
| `internal/parser` | Manifest load, PII tags, DFS, layer path matching |
| `internal/config` | YAML policy load, path normalization, CLI overrides |
| `internal/validator` | Violations, masking, confidential vs analysis policy |
| `cmd/dbt-guard` | Validate flag parsing, policy resolution |

Fixtures live in `internal/parser/testdata/` and `internal/config/testdata/`.

---

## E2E (CLI smoke tests)

Builds the binary and runs all commands against `examples/` and test manifests:

```bash
./scripts/test-e2e.sh
```

Checks:

1. PII column listing from `examples/`
2. `manifest` — declared PII IDs
3. `sensitive` — propagated sensitive nodes
4. `validate` — exit 1 on violation, exit 0 on masked fixture

Run before releases or after changes to `cmd/dbt-guard/main.go`.

---

## Manual testing (example project)

See [examples/README.md](../examples/README.md) for command-by-command expected output.

Quick validation with custom layer policy:

```bash
go build -o dbt-guard ./cmd/dbt-guard

./dbt-guard validate \
  internal/parser/testdata/manifest_with_confidential.json \
  --config examples/dbt-guard.yml
# expect: 1 violation (analysis_customers only)
```

---

## Real-world scenario (separate checkout)

Simulates consumption as another team would in production:

1. Build: `go build -o dbt-guard ./cmd/dbt-guard`
2. Copy `examples/` to a scratch dbt project (or use your own repo).
3. Add `dbt-guard.yml` with your warehouse layer paths.
4. Run `dbt compile` → `target/manifest.json`.
5. Gate: `dbt-guard validate target/manifest.json --config dbt-guard.yml`

Install the binary on `PATH` or invoke by absolute path in CI (GitHub Actions, GitLab, etc.).

---

## Debug (VS Code / Cursor)

Launch configuration **"Launch dbt-guard"** points at `examples/` by default (F5).
