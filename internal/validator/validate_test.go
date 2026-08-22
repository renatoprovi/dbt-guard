package validator

import (
	"path/filepath"
	"testing"

	"github.com/renatoprovi/dbt-guard/internal/config"
)

func TestRunValidate_Violation(t *testing.T) {
	path := filepath.Join("..", "parser", "testdata", "manifest_minimal.json")
	violations, err := RunValidate(path, config.DefaultLayerPolicy())
	if err != nil {
		t.Fatalf("RunValidate: %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	v := violations[0]
	if v.ModelID != "model.dbt_guard_example.analysis_customers" {
		t.Errorf("ModelID = %q", v.ModelID)
	}
	if len(v.LineagePath) != 3 {
		t.Errorf("expected 3 nodes in path, got %d: %v", len(v.LineagePath), v.LineagePath)
	}
	if v.LineagePath[0] != "model.dbt_guard_example.analysis_customers" ||
		v.LineagePath[1] != "model.dbt_guard_example.stg_customers" ||
		v.LineagePath[2] != "source.dbt_guard_example.raw.raw_customers" {
		t.Errorf("LineagePath = %v", v.LineagePath)
	}
}

func TestRunValidate_MaskedNoViolation(t *testing.T) {
	path := filepath.Join("..", "parser", "testdata", "manifest_analysis_masked.json")
	violations, err := RunValidate(path, config.DefaultLayerPolicy())
	if err != nil {
		t.Fatalf("RunValidate: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations (masked model), got %d: %v", len(violations), violations)
	}
}

func TestRunValidate_ConfidentialAllowed(t *testing.T) {
	path := filepath.Join("..", "parser", "testdata", "manifest_with_confidential.json")
	policy := config.LayerPolicy{
		PIIAllowed:    []string{"/confidential/"},
		PIIRestricted: []string{"/analysis/"},
	}.Normalize()
	violations, err := RunValidate(path, policy)
	if err != nil {
		t.Fatalf("RunValidate: %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation (analysis only), got %d: %v", len(violations), violations)
	}
	if violations[0].ModelID != "model.dbt_guard_example.analysis_customers" {
		t.Errorf("ModelID = %q", violations[0].ModelID)
	}
}

func TestRunValidate_FromConfigFile(t *testing.T) {
	manifestPath := filepath.Join("..", "parser", "testdata", "manifest_with_confidential.json")
	configPath := filepath.Join("..", "config", "testdata", "dbt-guard.yml")
	policy, err := config.LoadLayerPolicy(configPath)
	if err != nil {
		t.Fatalf("LoadLayerPolicy: %v", err)
	}
	violations, err := RunValidate(manifestPath, policy)
	if err != nil {
		t.Fatalf("RunValidate: %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation from config policy, got %d: %v", len(violations), violations)
	}
}
