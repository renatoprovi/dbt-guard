package parser

import (
	"path/filepath"
	"testing"
)

func TestLoadManifest(t *testing.T) {
	path := filepath.Join("testdata", "manifest_minimal.json")
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if m == nil {
		t.Fatal("LoadManifest returned nil")
	}
	if len(m.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(m.Nodes))
	}
	if len(m.Sources) != 1 {
		t.Errorf("expected 1 source, got %d", len(m.Sources))
	}
}

func TestLoadManifest_NotFound(t *testing.T) {
	_, err := LoadManifest("testdata/does_not_exist.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestSourceIDsWithPII(t *testing.T) {
	path := filepath.Join("testdata", "manifest_minimal.json")
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	ids := m.SourceIDsWithPII()
	if len(ids) != 1 {
		t.Fatalf("expected 1 source with PII, got %d: %v", len(ids), ids)
	}
	expected := "source.dbt_guard_example.raw.raw_clientes"
	if ids[0] != expected {
		t.Errorf("SourceIDsWithPII[0] = %q, want %q", ids[0], expected)
	}
}

func TestNodeIDsWithPII(t *testing.T) {
	path := filepath.Join("testdata", "manifest_minimal.json")
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	ids := m.NodeIDsWithPII()
	// In the fixture, no node has meta.security_tag: pii
	if len(ids) != 0 {
		t.Errorf("expected 0 nodes with PII, got %d: %v", len(ids), ids)
	}
}

func TestHasPIIColumn_ConfigMeta(t *testing.T) {
	s := &SourceDef{
		Columns: map[string]ColumnInfo{
			"email": {Config: &ConfigMeta{Meta: MetaMap{"security_tag": "pii"}}},
		},
	}
	if !s.HasPIIColumn() {
		t.Error("expected HasPIIColumn true for column with config.meta.security_tag: pii")
	}
}

func TestAnalysisNodeIDs(t *testing.T) {
	path := filepath.Join("testdata", "manifest_minimal.json")
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	ids := m.AnalysisNodeIDs()
	if len(ids) != 1 {
		t.Fatalf("expected 1 node in analysis, got %d: %v", len(ids), ids)
	}
	if ids[0] != "model.dbt_guard_example.analysis_clientes" {
		t.Errorf("AnalysisNodeIDs[0] = %q", ids[0])
	}
}

func TestIsNodeMasked(t *testing.T) {
	if IsNodeMasked(nil) {
		t.Error("nil must not be masked")
	}
	if !IsNodeMasked(&ManifestNode{Meta: MetaMap{"masked": true}}) {
		t.Error("meta.masked: true must be masked")
	}
	if !IsNodeMasked(&ManifestNode{Config: &ConfigMeta{Meta: MetaMap{"masked": true}}}) {
		t.Error("config.meta.masked: true must be masked")
	}
	if IsNodeMasked(&ManifestNode{Meta: MetaMap{}}) {
		t.Error("empty meta must not be masked")
	}
}
