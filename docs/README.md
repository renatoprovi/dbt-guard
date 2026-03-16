# Arquitetura e fluxo do dbt-guard

Este documento descreve a arquitetura do projeto e os fluxos de dados (atual e planejado).

---

## Visão geral

O dbt-guard é uma CLI de governança que usa **contrato declarativo** (YAML/JSON do dbt) para auditar linhagem e impedir que PII alcance camadas públicas sem mascaramento.

```mermaid
flowchart LR
    subgraph Entradas
        A[sources.yml]
        B[manifest.json]
    end
    subgraph dbt_guard["dbt-guard (Go)"]
        C[Parser]
        D[Validador]
    end
    subgraph Saídas
        E[Lista PII / Erros]
    end
    A --> C
    B --> C
    C --> D
    D --> E
```

---

## Arquitetura de componentes

```mermaid
flowchart TB
    subgraph CLI["cmd/dbt-guard"]
        M[main.go]
    end

    subgraph Internal["internal/"]
        subgraph Parser["parser/"]
            P1[sources.yml]
            P2[manifest.json]
            P3[FindSourceFiles]
            P4[ParseSourceFile]
            P5[LoadManifest]
            P6[CollectPIIColumns]
        end
        subgraph Validator["validator/"]
            V1[Regras de validação]
            V2[IsSensitive / DFS]
        end
    end

    M --> P3
    M --> P4
    M --> P6
    M --> V1
    P1 --> P4
    P2 --> P5
    P5 --> V2
    P4 --> P6
    V2 --> V1
```

| Componente | Responsabilidade |
|------------|------------------|
| **CLI** | Recebe pasta ou caminho do manifest, invoca parser e validador, imprime resultado ou sai com código de erro. |
| **Parser** | Lê `sources.yml` (recursivo) e `manifest.json`; expõe structs (SourceFile, Manifest, Node) e funções de busca/coleta de PII. |
| **Validator** | (Em evolução) Aplica regras: propagação de sensibilidade (DFS), checagem de mascaramento, validação da camada `analysis`. |

---

## Fluxo atual (sources.yml)

Hoje o dbt-guard percorre uma pasta, encontra todos os `sources.yml`, faz parse e imprime o nome das colunas com `security_tag: pii`.

```mermaid
sequenceDiagram
    participant U as Usuário
    participant CLI as CLI
    participant Find as FindSourceFiles
    participant Parse as ParseSourceFile
    participant Collect as CollectPIIColumns

    U->>CLI: dbt-guard ./caminho
    CLI->>Find: FindSourceFiles(root)
    Find-->>CLI: [paths...]
    loop Para cada sources.yml
        CLI->>Parse: ParseSourceFilePath(path)
        Parse-->>CLI: *SourceFile
        CLI->>Collect: CollectPIIColumns(path, sf)
        Collect-->>CLI: []PIIColumn
        CLI->>U: imprime nome da coluna
    end
```

```mermaid
flowchart LR
    A[Pasta] --> B[Busca sources.yml]
    B --> C[Parse YAML]
    C --> D[Filtra meta.security_tag: pii]
    D --> E[Imprime colunas]
```

---

## Fluxo alvo (manifest + validação)

Após o roadmap (Fase 1–3), o fluxo principal será: carregar o manifest, construir o grafo de linhagem, propagar PII por DFS e validar a camada `analysis`.

```mermaid
flowchart TB
    subgraph Entrada
        MF[manifest.json]
    end
    subgraph Carga
        LM[LoadManifest]
        G[Grafo de nodes + depends_on]
    end
    subgraph Propagação
        DFS[DFS a partir de cada nó]
        PII[IsSensitive?]
    end
    subgraph Validação
        AM[Modelos em analysis/]
        MASK[Possui mascaramento?]
        ERRO[Erro detalhado]
    end
    MF --> LM
    LM --> G
    G --> DFS
    DFS --> PII
    PII --> AM
    AM --> MASK
    MASK -->|Não mascarado| ERRO
```

```mermaid
sequenceDiagram
    participant U as Usuário
    participant CLI as validate
    participant Load as LoadManifest
    participant DFS as IsSensitive (DFS)
    participant Check as Checa analysis + mascaramento

    U->>CLI: dbt-guard validate --manifest target/manifest.json
    CLI->>Load: LoadManifest(path)
    Load-->>CLI: *Manifest
    loop Para cada modelo em analysis/
        CLI->>DFS: IsSensitive(nodeID, manifest)
        DFS-->>CLI: bool
        alt Sensível e não mascarado
            CLI->>Check: detalhes do caminho
            Check-->>CLI: erro
            CLI->>U: exit 1 + mensagem
        end
    end
```

---

## Grafo de linhagem (exemplo)

O manifest do dbt descreve um **grafo direcionado**: sources e modelos são nós; `depends_on` são arestas. A propagação de PII sobe das sources (onde está declarado no YAML) até os modelos que dependem delas.

```mermaid
flowchart LR
    subgraph Source["Source (declaração PII)"]
        S[raw.raw_clientes]
        S -->|cpf: security_tag pii| S
    end
    subgraph Staging
        ST[stg_clientes]
    end
    subgraph Analysis
        AN[analysis_clientes]
    end
    S -->|depends_on| ST
    ST -->|depends_on| AN
    AN -->|"documento_aluno (ex-cpf)"| AN
```

- **Source:** PII declarado em `sources.yml` (ex.: `cpf` com `meta.security_tag: pii`).
- **Staging:** depende da source; herda sensibilidade.
- **Analysis:** depende do staging; se expuser coluna PII sem tag de mascaramento, o validador deve falhar e reportar o caminho (ex.: `analysis_clientes` ← `stg_clientes` ← `raw.raw_clientes`).

---

## Camadas e regras de governança

```mermaid
flowchart TB
    subgraph Camadas
        direction TB
        R[Raw / Source]
        I[Intermediate / Staging]
        A[Analysis / Pública]
    end
    R -->|"PII declarado (meta)"| R
    R --> I
    I --> A
    A -->|"Não pode expor PII sem mascaramento"| G[Gatekeeper]
    G -->|OK| OK[exit 0]
    G -->|Violação| FAIL[exit 1 + caminho da linhagem]
```

| Camada | Papel | Regra |
|--------|--------|--------|
| **Source** | Contrato declarativo (sources.yml) | Colunas sensíveis com `meta.security_tag: pii`. |
| **Staging / Intermediate** | Refinamento; pode repassar PII para camadas internas. | — |
| **Analysis** | Dados expostos para consumo (relatórios, BI). | Não pode descender de PII sem estar explicitamente mascarado; caso contrário, o `validate` falha. |

---

## Resumo

| Artefato | Função |
|----------|--------|
| **sources.yml** | Declara quais colunas são PII (contrato). |
| **manifest.json** | Grafo de nós e `depends_on` (linhagem). |
| **Parser** | Lê YAML e JSON; expõe structs e listas de colunas PII. |
| **DFS / IsSensitive** | Propaga sensibilidade da source até o nó (grafo). |
| **validate** | Garante que modelos em `analysis/` que descendem de PII tenham mascaramento; senão, erro com caminho da linhagem. |

Os diagramas usam [Mermaid](https://mermaid.js.org/). Eles são renderizados no GitHub, no GitLab e em editores que suportam Mermaid (VS Code, Cursor com extensão).
