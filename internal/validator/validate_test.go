package validator

import (
	"path/filepath"
	"testing"
)

func TestRunValidate_Violation(t *testing.T) {
	path := filepath.Join("..", "parser", "testdata", "manifest_minimal.json")
	violations, err := RunValidate(path)
	if err != nil {
		t.Fatalf("RunValidate: %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("esperada 1 violação, obteve %d", len(violations))
	}
	v := violations[0]
	if v.ModelID != "model.dbt_guard_example.analysis_clientes" {
		t.Errorf("ModelID = %q", v.ModelID)
	}
	if len(v.LineagePath) != 3 {
		t.Errorf("esperado 3 nós no caminho, obteve %d: %v", len(v.LineagePath), v.LineagePath)
	}
	if v.LineagePath[0] != "model.dbt_guard_example.analysis_clientes" ||
		v.LineagePath[1] != "model.dbt_guard_example.stg_clientes" ||
		v.LineagePath[2] != "source.dbt_guard_example.raw.raw_clientes" {
		t.Errorf("LineagePath = %v", v.LineagePath)
	}
}

func TestRunValidate_MaskedNoViolation(t *testing.T) {
	path := filepath.Join("..", "parser", "testdata", "manifest_analysis_masked.json")
	violations, err := RunValidate(path)
	if err != nil {
		t.Fatalf("RunValidate: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("esperado 0 violações (modelo mascarado), obteve %d: %v", len(violations), violations)
	}
}
