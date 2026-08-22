#!/usr/bin/env bash
# dbt-guard E2E tests: build the binary and run all commands against examples/
# and internal/parser/testdata/, checking output and exit codes.
# Usage from repo root: ./scripts/test-e2e.sh

set -e
cd "$(dirname "$0")/.."
BIN="${BIN:-./dbt-guard}"
MANIFEST_MINIMAL="internal/parser/testdata/manifest_minimal.json"
MANIFEST_MASKED="internal/parser/testdata/manifest_analysis_masked.json"

echo "==> Building dbt-guard..."
go build -o "$BIN" ./cmd/dbt-guard

echo ""
echo "==> 1. Directory mode (PII columns in examples):"
out=$("$BIN" ./examples 2>&1) || true
if echo "$out" | grep -q "ssn"; then
  echo "    OK: output contains 'ssn'"
else
  echo "    FAILED: expected output containing 'ssn', got: $out"
  exit 1
fi

echo ""
echo "==> 2. manifest (IDs that declare PII):"
out=$("$BIN" manifest "$MANIFEST_MINIMAL" 2>&1) || true
if echo "$out" | grep -q "source.dbt_guard_example.raw.raw_customers"; then
  echo "    OK: PII source listed"
else
  echo "    FAILED: expected PII source in output, got: $out"
  exit 1
fi

echo ""
echo "==> 3. sensitive (sensitive nodes, DFS):"
out=$("$BIN" sensitive "$MANIFEST_MINIMAL" 2>&1) || true
for id in "source.dbt_guard_example.raw.raw_customers" "model.dbt_guard_example.stg_customers" "model.dbt_guard_example.analysis_customers"; do
  if echo "$out" | grep -q "$id"; then
    echo "    OK: $id"
  else
    echo "    FAILED: expected '$id' in output"
    exit 1
  fi
done

echo ""
echo "==> 4. validate without masking (must fail, exit 1):"
set +e
"$BIN" validate "$MANIFEST_MINIMAL" 2>&1; r=$?
set -e
if [ "$r" -eq 0 ]; then
  echo "    FAILED: validate should return exit 1 (violation)"
  exit 1
fi
echo "    OK: validate returned exit 1 (expected violation)"

echo ""
echo "==> 5. validate with masking (must pass, exit 0):"
if "$BIN" validate "$MANIFEST_MASKED" 2>&1; then
  echo "    OK: validate passed (exit 0)"
else
  echo "    FAILED: validate with masked model should return exit 0"
  exit 1
fi

echo ""
echo "==> All E2E tests passed."
