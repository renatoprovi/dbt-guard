package parser

import (
	"path/filepath"
	"testing"

	"github.com/renatocruz/dbt-guard/internal/config"
)

func TestRestrictedNodeIDs_DefaultPolicy(t *testing.T) {
	path := filepath.Join("testdata", "manifest_minimal.json")
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	ids := m.RestrictedNodeIDs(config.DefaultLayerPolicy())
	if len(ids) != 1 {
		t.Fatalf("expected 1 restricted node, got %d: %v", len(ids), ids)
	}
	if ids[0] != "model.dbt_guard_example.analysis_clientes" {
		t.Errorf("RestrictedNodeIDs[0] = %q", ids[0])
	}
}

func TestRestrictedNodeIDs_ConfidentialAllowed(t *testing.T) {
	path := filepath.Join("testdata", "manifest_with_confidential.json")
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	policy := config.LayerPolicy{
		PIIAllowed:    []string{"/confidential/"},
		PIIRestricted: []string{"/analysis/", "/confidential/"},
	}.Normalize()

	ids := m.RestrictedNodeIDs(policy)
	if len(ids) != 1 {
		t.Fatalf("expected only analysis restricted, got %d: %v", len(ids), ids)
	}
	if ids[0] != "model.dbt_guard_example.analysis_clientes" {
		t.Errorf("RestrictedNodeIDs[0] = %q", ids[0])
	}
}
