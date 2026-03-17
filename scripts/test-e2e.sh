#!/usr/bin/env bash
# Testes E2E do dbt-guard: compila o binário e executa todos os comandos
# contra examples/ e internal/parser/testdata/, verificando saída e exit code.
# Uso: da raiz do repo: ./scripts/test-e2e.sh

set -e
cd "$(dirname "$0")/.."
BIN="${BIN:-./dbt-guard}"
MANIFEST_MINIMAL="internal/parser/testdata/manifest_minimal.json"
MANIFEST_MASKED="internal/parser/testdata/manifest_analysis_masked.json"

echo "==> Compilando dbt-guard..."
go build -o "$BIN" ./cmd/dbt-guard

echo ""
echo "==> 1. Modo pasta (colunas PII em examples):"
out=$("$BIN" ./examples 2>&1) || true
if echo "$out" | grep -q "cpf"; then
  echo "    OK: saída contém 'cpf'"
else
  echo "    FALHOU: saída esperada contendo 'cpf', obteve: $out"
  exit 1
fi

echo ""
echo "==> 2. manifest (IDs que declaram PII):"
out=$("$BIN" manifest "$MANIFEST_MINIMAL" 2>&1) || true
if echo "$out" | grep -q "source.dbt_guard_example.raw.raw_clientes"; then
  echo "    OK: source com PII listada"
else
  echo "    FALHOU: esperado source PII na saída, obteve: $out"
  exit 1
fi

echo ""
echo "==> 3. sensitive (nós sensíveis DFS):"
out=$("$BIN" sensitive "$MANIFEST_MINIMAL" 2>&1) || true
for id in "source.dbt_guard_example.raw.raw_clientes" "model.dbt_guard_example.stg_clientes" "model.dbt_guard_example.analysis_clientes"; do
  if echo "$out" | grep -q "$id"; then
    echo "    OK: $id"
  else
    echo "    FALHOU: esperado '$id' na saída"
    exit 1
  fi
done

echo ""
echo "==> 4. validate sem mascaramento (deve falhar, exit 1):"
set +e
"$BIN" validate "$MANIFEST_MINIMAL" 2>&1; r=$?
set -e
if [ "$r" -eq 0 ]; then
  echo "    FALHOU: validate deveria retornar exit 1 (violação)"
  exit 1
fi
echo "    OK: validate retornou exit 1 (violação esperada)"

echo ""
echo "==> 5. validate com mascaramento (deve passar, exit 0):"
if "$BIN" validate "$MANIFEST_MASKED" 2>&1; then
  echo "    OK: validate passou (exit 0)"
else
  echo "    FALHOU: validate com masked deveria retornar exit 0"
  exit 1
fi

echo ""
echo "==> Todos os testes E2E passaram."
