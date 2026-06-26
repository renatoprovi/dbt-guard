# Example dbt project

Minimal **real** dbt project for testing dbt-guard governance (PII contract + lineage + layer policy). Main documentation: [README](../README.md).

## Model graph

| Layer | Path | Description |
|-------|------|-------------|
| Source | `models/sources.yml` | `raw_clientes`: `cpf` tagged `meta.security_tag: pii`, `nome` public. |
| Staging | `models/staging/stg_clientes.sql` | Selects from source; inherits PII sensitivity. |
| Analysis | `models/analysis/analysis_clientes.sql` | Exposes `cpf` as `documento_aluno` — **restricted** by default. |

`manifest.json` is **not** versioned. Generate it with `dbt compile`, or use repo fixtures under `internal/parser/testdata/`.

## Layer policy file

[dbt-guard.yml](dbt-guard.yml) shows a multi-zone warehouse policy:

```yaml
layers:
  pii_allowed:
    - "/models/raw_data/"
    - "/models/confidential/"
  pii_restricted:
    - "/models/analysis/"
```

This example project only has `staging/` and `analysis/` folders. The config illustrates how production teams map **confidential** (PII allowed) vs **analysis** (PII blocked).

---

## Test without dbt installed

From repository root (`dbt-guard/`):

```bash
# 1. PII columns from sources.yml
go run ./cmd/dbt-guard examples
# expected: cpf

# 2. Nodes/sources that declare PII
go run ./cmd/dbt-guard manifest internal/parser/testdata/manifest_minimal.json
# expected: source.dbt_guard_example.raw.raw_clientes

# 3. All sensitive nodes (DFS propagation)
go run ./cmd/dbt-guard sensitive internal/parser/testdata/manifest_minimal.json
# expected: source + stg_clientes + analysis_clientes

# 4. Validate — default policy (/analysis/ restricted)
go run ./cmd/dbt-guard validate internal/parser/testdata/manifest_minimal.json
# expected: exit 1 (analysis model, unmasked PII lineage)

# 4b. Validate — masked model passes
go run ./cmd/dbt-guard validate internal/parser/testdata/manifest_analysis_masked.json
# expected: exit 0

# 5. Validate — custom layer policy from config
go run ./cmd/dbt-guard validate \
  internal/parser/testdata/manifest_with_confidential.json \
  --config examples/dbt-guard.yml
# expected: exit 1 (analysis only; confidential is pii_allowed)
```

Fixture `manifest_with_confidential.json` adds a model under `models/confidential/` that also descends from PII — it must **not** violate when `confidential` is in `pii_allowed`.

---

## Test with dbt compile

Requires [dbt-core](https://docs.getdbt.com/docs/get-started/installation) (e.g. `pip install dbt-core dbt-postgres`):

```bash
cd examples
DBT_PROFILES_DIR=. dbt compile
cd ..
```

Then run against the compiled manifest:

```bash
go run ./cmd/dbt-guard manifest examples/target/manifest.json
go run ./cmd/dbt-guard sensitive examples/target/manifest.json
go run ./cmd/dbt-guard validate examples/target/manifest.json
go run ./cmd/dbt-guard validate examples/target/manifest.json --config examples/dbt-guard.yml
```

Postgres does not need to be running; `compile` only generates artifacts. `profiles.yml` targets Postgres for teams that also run `dbt run` locally.

---

## Command semantics (for governance reviews)

| Command | Question it answers |
|---------|---------------------|
| `manifest` | Who **declares** PII in the contract/manifest? |
| `sensitive` | Who **inherits** PII through lineage (direct + indirect)? |
| `validate` | Do **restricted-layer** models expose unmasked PII? (CI gate) |
