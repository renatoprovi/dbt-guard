# Projeto dbt de exemplo (dbt-guard)

Projeto dbt **mínimo** para testar a ferramenta de governança (PII e linhagem). Estrutura:

- **Source:** `models/sources.yml` — tabela `raw_clientes` com colunas `cpf` (meta.security_tag: pii) e `nome`.
- **Staging:** `models/staging/stg_clientes.sql` — `SELECT *` da source.
- **Analysis:** `models/analysis/analysis_clientes.sql` — `SELECT cpf AS documento_aluno, nome FROM ref('stg_clientes')`.

O `manifest.json` gerado por `dbt compile` contém o grafo de dependências (`depends_on`) entre source → staging → analysis, usado pelo validador em Go para percorrer a linhagem.

## Como compilar (gerar manifest.json)

Na pasta **examples** (raiz do projeto dbt):

```bash
cd examples
DBT_PROFILES_DIR=. dbt compile
```

O manifest fica em `examples/target/manifest.json`.

Requisitos: [dbt-core](https://docs.getdbt.com/docs/get-started/installation) instalado (ex.: `pip install dbt-core dbt-postgres`). O `profiles.yml` usa Postgres; para apenas compilar e gerar o manifest, não é obrigatório ter o banco rodando em todas as versões do dbt.
