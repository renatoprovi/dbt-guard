# Roadmap do dbt-guard

## Fase 1: Estruturas (grafo) — Implementado

Parser do `manifest.json`, `NodeIDsWithPII`, `SourceIDsWithPII`. Comando: `dbt-guard manifest <path>`.

## Fase 2: DFS — Implementado

`IsSensitive`, propagação de sensibilidade pelo grafo. Comando: `dbt-guard sensitive <path>`.

## Fase 3: Validate — Implementado

`RunValidate`, checagem de modelos em `analysis/` e mascaramento (`meta.masked: true`). Comando: `dbt-guard validate <path>`.
