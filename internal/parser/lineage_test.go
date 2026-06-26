package parser

import (
	"path/filepath"
	"testing"
)

func TestIsSensitive_SourceWithPII(t *testing.T) {
	path := filepath.Join("testdata", "manifest_minimal.json")
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	sourceID := "source.dbt_guard_example.raw.raw_clientes"
	if !IsSensitive(sourceID, m) {
		t.Errorf("IsSensitive(%q) = false, want true (source has PII column)", sourceID)
	}
}

func TestIsSensitive_ModelDescendsFromPII(t *testing.T) {
	path := filepath.Join("testdata", "manifest_minimal.json")
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	// stg_clientes depends on source raw.raw_clientes which has PII
	stgID := "model.dbt_guard_example.stg_clientes"
	if !IsSensitive(stgID, m) {
		t.Errorf("IsSensitive(%q) = false, want true (descends from PII source)", stgID)
	}
	// analysis_clientes depends on stg_clientes -> PII source
	analysisID := "model.dbt_guard_example.analysis_clientes"
	if !IsSensitive(analysisID, m) {
		t.Errorf("IsSensitive(%q) = false, want true (descends from PII source)", analysisID)
	}
}

func TestIsSensitive_UnknownID(t *testing.T) {
	path := filepath.Join("testdata", "manifest_minimal.json")
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if IsSensitive("model.fake.xyz", m) {
		t.Error("IsSensitive(unknown) = true, want false")
	}
}

func TestIsSensitive_NilManifest(t *testing.T) {
	if IsSensitive("model.x.y", nil) {
		t.Error("IsSensitive with nil manifest must return false")
	}
}

func TestLineagePathToPII(t *testing.T) {
	path := filepath.Join("testdata", "manifest_minimal.json")
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	p := LineagePathToPII("model.dbt_guard_example.analysis_clientes", m)
	if len(p) != 3 {
		t.Fatalf("expected path with 3 nodes, got %d: %v", len(p), p)
	}
	if p[0] != "model.dbt_guard_example.analysis_clientes" ||
		p[1] != "model.dbt_guard_example.stg_clientes" ||
		p[2] != "source.dbt_guard_example.raw.raw_clientes" {
		t.Errorf("LineagePathToPII = %v", p)
	}
	if len(LineagePathToPII("model.fake.xyz", m)) != 0 {
		t.Error("unknown node must return empty path")
	}
}
