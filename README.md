# dbt-guard

Ferramenta de governança em Go que audita automaticamente o grafo de linhagem do dbt para impedir que dados sensíveis (Personally Identifiable Information - PII) alcancem camadas públicas de análise sem o devido mascaramento, garantindo conformidade com a LGPD através de um contrato de dados declarativo.

## Pré-requisitos

- [Go 1.22+](https://go.dev/dl/)

## Estrutura do projeto

```
dbt-guard/
├── cmd/dbt-guard/       # Entrada do programa (main.go)
├── internal/
│   ├── parser/          # Parse de sources.yml e detecção de colunas PII
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
go run ./cmd/dbt-guard [pasta]
```

**Compilar um binário:**

```bash
go build -o dbt-guard ./cmd/dbt-guard
./dbt-guard [pasta]
```

O argumento `[pasta]` é o diretório a partir do qual o dbt-guard procura recursivamente por arquivos `sources.yml`. Se omitido, usa o diretório atual (`.`).

## Como testar

1. **Usando a pasta de exemplos** (recomendado para validar o comportamento):

   ```bash
   go run ./cmd/dbt-guard ./examples
   ```

   Saída esperada (nomes das colunas com `security_tag: pii`):

   ```
   email
   full_name
   customer_email
   taxpayer_id
   ```

2. **Testar com seu próprio projeto dbt:**  
   Passe o caminho da pasta onde estão seus `models` (ou a raiz do projeto):

   ```bash
   go run ./cmd/dbt-guard /caminho/para/seu/projeto/dbt
   
   go run ./cmd/dbt-guard examples/models
   ```

3. **No Cursor/VS Code:**  
   Use a configuração de debug **"Launch dbt-guard"** (F5). Ela já aponta para a pasta `examples` por padrão.

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

- **Testes:** `go test ./...`
- **Build de todos os pacotes:** `go build ./...`
- **Debug:** use o launch **"Launch dbt-guard"** no VS Code/Cursor (F5).

---

## Roadmap

### Fase 1: Definindo as Estruturas de Dados (O Grafo)

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

---

### Fase 2: Implementando o Motor de Propagação (DFS)

Implementar a **propagação de sensibilidade** na linhagem:

- Criar a função **`IsSensitive(nodeID string, manifest Manifest) bool`**.
- Percorrer **recursivamente o grafo** a partir de um nó destino até chegar em uma **source**.
- Se **qualquer parent** na linhagem tiver **`security_tag: pii`** no meta, o nó atual deve ser considerado PII.
- Usar **DFS (Busca em Profundidade)** e otimizar o uso de memória (evitar visitar o mesmo nó múltiplas vezes, evitar estruturas desnecessárias).

---

### Fase 3: O Validador de PRs (Gatekeeper)

Criar o **comando CLI `validate`**:

1. **Ler o manifest.json.**
2. **Identificar todos os modelos na pasta `analysis`.**
3. Para cada modelo em `analysis`:
   - Verificar se ele **descende de alguma fonte PII** (usar o motor da Fase 2).
   - Se descender de PII e **não** estiver explicitamente mascarado (verificar colunas com tags de mascaramento), **retornar erro detalhado** listando:
     - o modelo em questão,
     - o caminho da linhagem que causou o alerta.

Objetivo: impedir que modelos na camada `analysis` exponham PII sem mascaramento.

---

## Contribuição

- **Bugs e ideias:** abra uma issue descrevendo o problema ou a sugestão.
- **Código:** envie um pull request a partir da `main`. Garanta que `go test ./...` e `go build ./...` passem.
- **Padrão:** use `gofmt` e as regras de lint do projeto (ex.: staticcheck).
- **Repositório:** ao publicar (GitHub/GitLab etc.), indique o link na descrição do projeto.

---

## Licença

Projeto em desenvolvimento; use conforme sua política interna.
