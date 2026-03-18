# dbt-guard

Ferramenta de governança em Go que audita automaticamente o grafo de linhagem do dbt para impedir que dados sensíveis (Personally Identifiable Information - PII) alcancem camadas públicas de análise sem o devido mascaramento, garantindo conformidade com a LGPD através de um contrato de dados declarativo.

## Pré-requisitos

- [Go 1.22+](https://go.dev/dl/)

## Estrutura do projeto

```
dbt-guard/
├── cmd/dbt-guard/       # Entrada do programa (main.go)
├── internal/
│   ├── parser/          # Parse de sources.yml + manifest.json (linhagem)
│   └── validator/       # (futuro) Regras de validação
├── examples/            # Projeto dbt mínimo (source → staging → analysis)
│   ├── dbt_project.yml
│   ├── profiles.yml
│   ├── models/
│   │   ├── sources.yml      # raw_clientes (cpf PII, nome)
│   │   ├── staging/
│   │   │   └── stg_clientes.sql
│   │   └── analysis/
│   │       └── analysis_clientes.sql
│   └── target/              # manifest.json após dbt compile
├── go.mod
└── README.md
```

## Instalação e execução

**Compilar e rodar (sem instalar):**

```bash
# Listar colunas PII a partir de sources.yml (busca recursiva na pasta)
go run ./cmd/dbt-guard [pasta]

# Listar IDs de nós/sources com PII declarado no manifest (Fase 1)
go run ./cmd/dbt-guard manifest <caminho/manifest.json>

# Listar IDs sensíveis (propagação DFS): nós que descendem de PII (Fase 2)
go run ./cmd/dbt-guard sensitive <caminho/manifest.json>

# Validar camada analysis: falha se algum modelo em analysis/ descender de PII sem mascaramento (Fase 3)
go run ./cmd/dbt-guard validate <caminho/manifest.json>
```

**Compilar um binário:**

```bash
go build -o dbt-guard ./cmd/dbt-guard
./dbt-guard [pasta]
./dbt-guard manifest <caminho/manifest.json>
./dbt-guard sensitive <caminho/manifest.json>
./dbt-guard validate <caminho/manifest.json>
```

- **Modo pasta:** procura recursivamente por `sources.yml` e imprime o nome das colunas com `security_tag: pii`.
- **Modo manifest:** carrega o `manifest.json`, imprime os `unique_id` de nodes/sources que **declaram** PII (meta ou colunas).
- **Modo sensitive:** carrega o `manifest.json`, percorre o grafo em DFS e imprime os `unique_id` de todos os nós/sources **sensíveis** (que declaram PII ou descendem de um que declara).
- **Modo validate:** carrega o `manifest.json`, lista modelos em `analysis/`; se algum **descender de PII** e **não** tiver `meta.masked: true` (ou `config.meta.masked`), imprime erro com o modelo e o caminho da linhagem e sai com código 1.

## Como testar

1. **Usando a pasta de exemplos** (recomendado para validar o comportamento):

   ```bash
   go run ./cmd/dbt-guard ./examples
   ```

   Saída esperada (nomes das colunas com `security_tag: pii`):

   ```
   cpf
   ```

2. **Testar com seu próprio projeto dbt:**  
   Passe o caminho da pasta onde estão seus `models` (ou a raiz do projeto):

   ```bash
   go run ./cmd/dbt-guard /caminho/para/seu/projeto/dbt
   
   go run ./cmd/dbt-guard examples/models
   ```

3. **Testar o comando manifest (grafo de linhagem):**  
   Use um manifest gerado por `dbt compile` ou o fixture de testes:

   ```bash
   go run ./cmd/dbt-guard manifest internal/parser/testdata/manifest_minimal.json
   ```
   Saída esperada: um `unique_id` de source com coluna PII (ex.: `source.dbt_guard_example.raw.raw_clientes`).

4. **Testar propagação DFS (Fase 2 — sensitive):**
   ```bash
   go run ./cmd/dbt-guard sensitive internal/parser/testdata/manifest_minimal.json
   ```
   Saída esperada: source + os dois models que dependem dela (todos sensíveis):
   ```
   model.dbt_guard_example.stg_clientes
   model.dbt_guard_example.analysis_clientes
   source.dbt_guard_example.raw.raw_clientes
   ```

5. **Testar validação da camada analysis (Fase 3 — validate):**
   ```bash
   # Sem mascaramento: deve falhar com violação
   go run ./cmd/dbt-guard validate internal/parser/testdata/manifest_minimal.json
   # Com meta.masked: true no modelo analysis: deve passar (exit 0)
   go run ./cmd/dbt-guard validate internal/parser/testdata/manifest_analysis_masked.json
   ```
   Na violação, a saída mostra o modelo e a linhagem (ex.: `analysis_clientes → stg_clientes → source.raw.raw_clientes`).

6. **No Cursor/VS Code:**  
   Use a configuração de debug **"Launch dbt-guard"** (F5). Ela já aponta para a pasta `examples` por padrão.

### Testar efetivamente (estratégia completa)

- **Testes unitários:** na raiz do repo, rode `go test ./...`. Eles cobrem o parser (manifest, linhagem, PII) e o validator (violações e mascaramento) usando os fixtures em `internal/parser/testdata/`.
- **Testes E2E do binário:** use o script que compila o binário e executa todos os comandos contra `examples/` e os manifests de teste, conferindo saída e exit code:
  ```bash
  ./scripts/test-e2e.sh
  ```
  Útil para validar o CLI de ponta a ponta antes de release ou após mudanças no `main`.

### Testar em cenário real (repositório separado + binário)

Para ver como o dbt-guard se comporta **como se fosse usado em outro repositório** (binário instalado ou em PATH, projeto dbt à parte):

1. **Compilar o binário** (no clone do dbt-guard):
   ```bash
   cd /caminho/para/dbt-guard
   go build -o dbt-guard ./cmd/dbt-guard
   ```

2. **Criar um “repositório de teste”** com um projeto dbt (pode ser cópia do `examples` ou seu próprio projeto):
   ```bash
   mkdir -p ~/dbt-guard-cenario-real
   cp -r /caminho/para/dbt-guard/examples/* ~/dbt-guard-cenario-real/
   cd ~/dbt-guard-cenario-real
   ```

3. **Gerar o manifest** no projeto de teste (exige [dbt-core](https://docs.getdbt.com/docs/get-started/installation) instalado, ex.: `pip install dbt-core dbt-postgres`):
   ```bash
   DBT_PROFILES_DIR=. dbt compile
   ```
   O banco não precisa estar rodando; o `compile` só gera `target/manifest.json`.

4. **Usar o binário do dbt-guard** a partir do projeto de teste. Duas opções:
   - **Binário no PATH:** copie para um diretório no PATH (ex.: `cp /caminho/para/dbt-guard/dbt-guard ~/go/bin/` ou `sudo cp ... /usr/local/bin/`) e rode:
     ```bash
     cd ~/dbt-guard-cenario-real
     dbt-guard .                                    # colunas PII nos YAML
     dbt-guard manifest target/manifest.json
     dbt-guard sensitive target/manifest.json
     dbt-guard validate target/manifest.json        # deve falhar (analysis sem mascaramento)
     ```
   - **Binário por caminho absoluto:** sem instalar no PATH:
     ```bash
     cd ~/dbt-guard-cenario-real
     /caminho/para/dbt-guard/dbt-guard .
     /caminho/para/dbt-guard/dbt-guard manifest target/manifest.json
     /caminho/para/dbt-guard/dbt-guard sensitive target/manifest.json
     /caminho/para/dbt-guard/dbt-guard validate target/manifest.json
     ```

5. **Resultado esperado neste cenário:**  
   - `dbt-guard .` → imprime `cpf` (coluna PII do `sources.yml`).  
   - `manifest` → imprime o `unique_id` da source com PII.  
   - `sensitive` → imprime source + modelos que dependem dela.  
   - `validate` → **exit 1** e mensagem de violação (modelo em `analysis/` descende de PII sem mascaramento). Para fazer passar, adicione `meta: { masked: true }` no modelo em `models/analysis/` e rode `dbt compile` de novo; depois `dbt-guard validate target/manifest.json` deve retornar exit 0.

Assim você valida o comportamento do binário exatamente como em um uso real (outro repo, CI ou máquina com o binário instalado).

## Formato YAML (sources do dbt)

O dbt-guard procura arquivos chamados **`sources.yml`** e, em cada um, lê a estrutura de `sources` → `tables` → `columns`. Uma coluna é considerada PII se tiver:

- **`meta.security_tag: pii`**, ou  
- **`config.meta.security_tag: pii`** (estilo dbt v1.10+)

Exemplo mínimo em uma coluna:

```yaml
columns:
  - name: email
    meta:
      security_tag: pii
```

## Documentação

- **[Arquitetura e fluxo](docs/README.md)** — diagramas da arquitetura, fluxo atual (sources.yml), fluxo alvo (manifest + validate) e grafo de linhagem.

## Desenvolvimento

- **Testes unitários:** `go test ./...`
- **Testes E2E (binário):** `./scripts/test-e2e.sh`
- **Build de todos os pacotes:** `go build ./...`
- **Debug:** use o launch **"Launch dbt-guard"** no VS Code/Cursor (F5).

---

## Roadmap

### Fase 1: Definindo as Estruturas de Dados (O Grafo) ✅ Implementado

Construir a estrutura de dados do dbt-guard focada em **análise de linhagem**:

- Definir **structs em Go** que façam unmarshal do **manifest.json** do dbt (versão 10 ou superior).
- Mapear os campos **`nodes`** (modelos e fontes) e suas dependências (**`depends_on`**).
- Implementar uma função que **carregue o manifest.json** nessas structs.

**Passo 1 — Estrutura do manifest (instrução para implementação):**

- Criar o arquivo **`internal/parser/manifest.go`** com:
  - Uma struct **`Manifest`** que contenha um **mapa de nodes** (chave = ID do nó).
  - Uma struct **`Node`** com: `UniqueId`, `ResourceType` (ex.: `"model"`, `"source"`), **`DependsOn`** (lista de nós pais) e **`Meta`** (tags de segurança).
  - Função **`LoadManifest(path string) (*Manifest, error)`** usando `os.ReadFile` e `json.Unmarshal`. Usar **`json.RawMessage`** para campos que não serão processados agora (otimização de memória).
- **Por que structs com tags (`json:"nodes"`) em Go:** são mais performáticas e seguras que `map[string]interface{}` porque evitam alocações dinâmicas, permitem tipagem forte e acesso direto aos campos; o decoder do Go mapeia só o que as structs declaram.

**Passo 2 — Dicionário declarativo (preparação no repositório dbt):**

- No projeto dbt, escolher **uma source** com dados sensíveis e declarar as tags no **`sources.yml`**:

```yaml
version: 2
sources:
  - name: erp_origem
    tables:
      - name: clientes
        columns:
          - name: cpf
            meta:
              security_tag: pii
          - name: nome_completo
            meta:
              security_tag: pii
          - name: data_cadastro
            meta:
              security_tag: public
```

- **Objetivo:** validação declarativa (o YAML é o “contrato”); o dbt-guard usa o manifest para percorrer o grafo sem depender de nomes de arquivo.
- **Critério de conclusão da Fase 1:** carregar o manifest, imprimir os IDs dos nós que possuem a tag `pii`; em seguida partir para a busca recursiva (DFS).  
  **Status:** implementado em `internal/parser/manifest.go` (`LoadManifest`, `Manifest`, `ManifestNode`, `SourceDef`, `NodeIDsWithPII`, `SourceIDsWithPII`). Comando: `dbt-guard manifest <path>`. Testes em `internal/parser/manifest_test.go` e fixture em `internal/parser/testdata/manifest_minimal.json`.

---

### Fase 2: Implementando o Motor de Propagação (DFS) ✅ Implementado

Implementar a **propagação de sensibilidade** na linhagem:

- Criar a função **`IsSensitive(nodeID string, manifest Manifest) bool`**.
- Percorrer **recursivamente o grafo** a partir de um nó destino até chegar em uma **source**.
- Se **qualquer parent** na linhagem tiver **`security_tag: pii`** no meta (ou for source com coluna PII), o nó atual deve ser considerado PII.
- Usar **DFS (Busca em Profundidade)** e otimizar o uso de memória (evitar visitar o mesmo nó múltiplas vezes, evitar estruturas desnecessárias).

**Status:** implementado em `internal/parser/lineage.go` (`IsSensitive` com DFS e cache por nodeID). Comando `dbt-guard sensitive <path>` imprime todos os nós e sources sensíveis. Testes em `internal/parser/lineage_test.go`.

---

### Fase 3: O Validador de PRs (Gatekeeper) ✅ Implementado

Criar o **comando CLI `validate`**:

1. **Ler o manifest.json.**
2. **Identificar todos os modelos na pasta `analysis`.**
3. Para cada modelo em `analysis`:
   - Verificar se ele **descende de alguma fonte PII** (usar o motor da Fase 2).
   - Se descender de PII e **não** estiver explicitamente mascarado (verificar colunas com tags de mascaramento), **retornar erro detalhado** listando:
     - o modelo em questão,
     - o caminho da linhagem que causou o alerta.

Objetivo: impedir que modelos na camada `analysis` exponham PII sem mascaramento.

**Status:** implementado em `internal/validator/validate.go` (`RunValidate`, `Violation`) e parser (`AnalysisNodeIDs`, `LineagePathToPII`, `IsNodeMasked`). Comando `dbt-guard validate <path>`: exit 1 com mensagem e linhagem em caso de violação. Mascaramento via `meta.masked: true` ou `config.meta.masked: true` no modelo. Testes em `internal/validator/validate_test.go` e fixtures `manifest_minimal.json` / `manifest_analysis_masked.json`.

---

## Contribuição

- **Bugs e ideias:** abra uma issue descrevendo o problema ou a sugestão.
- **Código:** envie um pull request a partir da `main`. Garanta que `go test ./...` e `go build ./...` passem.
- **Padrão:** use `gofmt` e as regras de lint do projeto (ex.: staticcheck).
- **Repositório:** ao publicar (GitHub/GitLab etc.), indique o link na descrição do projeto.

---

## Licença

Projeto em desenvolvimento; use conforme sua política interna.
