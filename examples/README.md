# Projeto dbt de exemplo (dbt-guard)

Projeto dbt **real** e mínimo para testar a ferramenta de governança (PII e linhagem). Visão geral do dbt-guard e comandos: [README principal](../README.md). Estrutura:

- **Source:** `models/sources.yml` — tabela `raw_clientes` com colunas `cpf` (meta.security_tag: pii) e `nome`.
- **Staging:** `models/staging/stg_clientes.sql` — `SELECT *` da source.
- **Analysis:** `models/analysis/analysis_clientes.sql` — `SELECT cpf AS documento_aluno, nome FROM ref('stg_clientes')`.

O **manifest** não vem versionado: é gerado pelo próprio dbt ao rodar `dbt compile` nesta pasta. Se você **não tem dbt instalado**, use os manifests de teste do repositório (veja abaixo).

---

## Testar o dbt-guard (sem precisar do dbt)

Na **raiz do repositório** (`dbt-guard/`). O comando 1 usa a pasta `examples`; os comandos 2–4 usam manifests de teste que já vêm no repo (mesmo grafo do examples).

```bash
# 1. Colunas PII nos YAML (só precisa da pasta examples)
go run ./cmd/dbt-guard examples
# → cpf

# 2. IDs que declaram PII no manifest
go run ./cmd/dbt-guard manifest internal/parser/testdata/manifest_minimal.json
# → source.dbt_guard_example.raw.raw_clientes

# 3. Nós sensíveis (DFS)
go run ./cmd/dbt-guard sensitive internal/parser/testdata/manifest_minimal.json
# → 3 IDs (source + stg_clientes + analysis_clientes)

# 4. Validate (deve falhar: analysis sem mascaramento)
go run ./cmd/dbt-guard validate internal/parser/testdata/manifest_minimal.json
# → exit 1 + mensagem de violação

# 4b. Validate com modelo mascarado (deve passar)
go run ./cmd/dbt-guard validate internal/parser/testdata/manifest_analysis_masked.json
# → exit 0
```

---

## Testar com o manifest gerado pelo dbt (com dbt instalado)

Se você tem [dbt-core](https://docs.getdbt.com/docs/get-started/installation) instalado (ex.: `pip install dbt-core dbt-postgres`), pode gerar o manifest a partir deste projeto:

```bash
cd examples
DBT_PROFILES_DIR=. dbt compile
cd ..
```

Depois use `examples/target/manifest.json` nos comandos:

```bash
go run ./cmd/dbt-guard manifest examples/target/manifest.json
go run ./cmd/dbt-guard sensitive examples/target/manifest.json
go run ./cmd/dbt-guard validate examples/target/manifest.json
```

O `profiles.yml` usa Postgres; para apenas compilar, o banco não precisa estar rodando.
