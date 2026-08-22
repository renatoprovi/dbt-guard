package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultLayerPolicy(t *testing.T) {
	p := DefaultLayerPolicy()
	if !p.IsPIIRestricted("models/analysis/report.sql") {
		t.Error("expected /analysis/ to be restricted by default")
	}
	if p.IsPIIAllowed("models/analysis/report.sql") {
		t.Error("analysis must not be allowed by default")
	}
}

func TestLoadLayerPolicy(t *testing.T) {
	path := filepath.Join("testdata", "dbt-guard.yml")
	p, err := LoadLayerPolicy(path)
	if err != nil {
		t.Fatalf("LoadLayerPolicy: %v", err)
	}
	if !p.IsPIIAllowed("models/confidential/finance.sql") {
		t.Error("confidential should be allowed")
	}
	if !p.IsPIIRestricted("models/analysis/report.sql") {
		t.Error("analysis should be restricted")
	}
	if p.IsPIIRestricted("models/dwh/dim_customer.sql") {
		t.Error("dwh should be neutral (not restricted)")
	}
}

func TestWithCLIOverrides(t *testing.T) {
	p := DefaultLayerPolicy().WithCLIOverrides([]string{"confidential"}, nil)
	if !p.IsPIIAllowed("models/confidential/finance.sql") {
		t.Error("CLI --allowed confidential should allow path")
	}
	if !p.IsPIIRestricted("models/analysis/report.sql") {
		t.Error("default restricted should remain")
	}
}

func TestAllowedWinsOverRestricted(t *testing.T) {
	p := LayerPolicy{
		PIIAllowed:    []string{"/confidential/"},
		PIIRestricted: []string{"/confidential/", "/analysis/"},
	}.Normalize()
	path := "models/confidential/finance.sql"
	if !p.IsPIIAllowed(path) {
		t.Fatal("path should match allowed")
	}
	// RestrictedNodeIDs logic skips allowed first; IsPIIRestricted alone may still be true
	if !p.IsPIIRestricted(path) {
		t.Error("path also matches restricted pattern")
	}
}

func TestNormalizePathPattern(t *testing.T) {
	cases := map[string]string{
		"analysis":         "/analysis/",
		"/models/analysis": "/models/analysis/",
		"confidential/":    "/confidential/",
	}
	for in, want := range cases {
		if got := normalizePathPattern(in); got != want {
			t.Errorf("normalizePathPattern(%q) = %q, want %q", in, got, want)
		}
	}
}
