package main

import (
	"testing"
)

func TestParseValidateArgs(t *testing.T) {
	opts, err := parseValidateArgs([]string{"target/manifest.json"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.ManifestPath != "target/manifest.json" {
		t.Errorf("ManifestPath = %q", opts.ManifestPath)
	}
}

func TestParseValidateArgs_Flags(t *testing.T) {
	opts, err := parseValidateArgs([]string{
		"--allowed", "confidential, raw_data",
		"--restricted", "analysis",
		"target/manifest.json",
	})
	if err != nil {
		t.Fatal(err)
	}
	if opts.ConfigPath != "" {
		t.Errorf("ConfigPath = %q", opts.ConfigPath)
	}
	if len(opts.Allowed) != 2 {
		t.Errorf("Allowed = %v", opts.Allowed)
	}
	policy, err := resolveLayerPolicy(opts)
	if err != nil {
		t.Fatal(err)
	}
	if !policy.IsPIIAllowed("models/confidential/x.sql") {
		t.Error("expected confidential allowed")
	}
	if !policy.IsPIIRestricted("models/analysis/x.sql") {
		t.Error("expected analysis restricted")
	}
}

func TestResolveLayerPolicy_Default(t *testing.T) {
	policy, err := resolveLayerPolicy(validateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !policy.IsPIIRestricted("models/analysis/x.sql") {
		t.Error("expected default /analysis/ restricted")
	}
}

func TestParseValidateArgs_MissingManifest(t *testing.T) {
	_, err := parseValidateArgs([]string{"--config", "x.yml"})
	if err == nil {
		t.Fatal("expected error when manifest path missing")
	}
}
