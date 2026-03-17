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
		t.Fatal("LoadManifest retornou nil")
	}
	if len(m.Nodes) != 2 {
		t.Errorf("esperado 2 nodes, obteve %d", len(m.Nodes))
	}
	if len(m.Sources) != 1 {
		t.Errorf("esperado 1 source, obteve %d", len(m.Sources))
	}
}

func TestLoadManifest_NotFound(t *testing.T) {
	_, err := LoadManifest("testdata/nao_existe.json")
	if err == nil {
		t.Fatal("esperado erro para arquivo inexistente")
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
		t.Fatalf("esperada 1 source com PII, obteve %d: %v", len(ids), ids)
	}
	expected := "source.dbt_guard_example.raw.raw_clientes"
	if ids[0] != expected {
		t.Errorf("SourceIDsWithPII[0] = %q, esperado %q", ids[0], expected)
	}
}

func TestNodeIDsWithPII(t *testing.T) {
	path := filepath.Join("testdata", "manifest_minimal.json")
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	ids := m.NodeIDsWithPII()
	// No fixture, nenhum node tem meta.security_tag: pii
	if len(ids) != 0 {
		t.Errorf("esperado 0 nodes com PII, obteve %d: %v", len(ids), ids)
	}
}

func TestHasPIIColumn_ConfigMeta(t *testing.T) {
	s := &SourceDef{
		Columns: map[string]ColumnInfo{
			"email": {Config: &ConfigMeta{Meta: MetaMap{"security_tag": "pii"}}},
		},
	}
	if !s.HasPIIColumn() {
		t.Error("esperado HasPIIColumn true para coluna com config.meta.security_tag: pii")
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
		t.Fatalf("esperado 1 nó em analysis, obteve %d: %v", len(ids), ids)
	}
	if ids[0] != "model.dbt_guard_example.analysis_clientes" {
		t.Errorf("AnalysisNodeIDs[0] = %q", ids[0])
	}
}

func TestIsNodeMasked(t *testing.T) {
	if IsNodeMasked(nil) {
		t.Error("nil não deve ser mascarado")
	}
	if !IsNodeMasked(&ManifestNode{Meta: MetaMap{"masked": true}}) {
		t.Error("meta.masked: true deve ser mascarado")
	}
	if !IsNodeMasked(&ManifestNode{Config: &ConfigMeta{Meta: MetaMap{"masked": true}}}) {
		t.Error("config.meta.masked: true deve ser mascarado")
	}
	if IsNodeMasked(&ManifestNode{Meta: MetaMap{}}) {
		t.Error("meta vazio não deve ser mascarado")
	}
}
