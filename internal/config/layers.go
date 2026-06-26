package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// LayerPolicy defines which model paths may or must not carry unmasked PII lineage.
type LayerPolicy struct {
	PIIAllowed    []string `yaml:"pii_allowed"`
	PIIRestricted []string `yaml:"pii_restricted"`
}

type file struct {
	Layers LayerPolicy `yaml:"layers"`
}

// DefaultLayerPolicy matches the original behavior: only models under /analysis/ are gated.
func DefaultLayerPolicy() LayerPolicy {
	return LayerPolicy{
		PIIRestricted: []string{"/analysis/"},
	}
}

// LoadLayerPolicy reads dbt-guard.yml (or any YAML file with a top-level "layers" key).
func LoadLayerPolicy(path string) (LayerPolicy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return LayerPolicy{}, err
	}
	var f file
	if err := yaml.Unmarshal(data, &f); err != nil {
		return LayerPolicy{}, err
	}
	return f.Layers.normalized(), nil
}

// WithCLIOverrides appends CLI --allowed and --restricted patterns to the policy.
func (p LayerPolicy) WithCLIOverrides(allowed, restricted []string) LayerPolicy {
	out := p.normalized()
	for _, a := range allowed {
		if a = normalizePathPattern(a); a != "" {
			out.PIIAllowed = append(out.PIIAllowed, a)
		}
	}
	for _, r := range restricted {
		if r = normalizePathPattern(r); r != "" {
			out.PIIRestricted = append(out.PIIRestricted, r)
		}
	}
	return out.normalized()
}

// Normalize returns a copy with normalized path patterns.
func (p LayerPolicy) Normalize() LayerPolicy {
	return p.normalized()
}

func (p LayerPolicy) normalized() LayerPolicy {
	var allowed, restricted []string
	for _, s := range p.PIIAllowed {
		if s = normalizePathPattern(s); s != "" {
			allowed = append(allowed, s)
		}
	}
	for _, s := range p.PIIRestricted {
		if s = normalizePathPattern(s); s != "" {
			restricted = append(restricted, s)
		}
	}
	return LayerPolicy{PIIAllowed: allowed, PIIRestricted: restricted}
}

// IsPIIAllowed reports whether original_file_path is in an allowed layer (PII permitted without masking).
func (p LayerPolicy) IsPIIAllowed(filePath string) bool {
	for _, pattern := range p.PIIAllowed {
		if pathMatches(filePath, pattern) {
			return true
		}
	}
	return false
}

// IsPIIRestricted reports whether original_file_path is in a restricted layer (unmasked PII lineage fails validate).
func (p LayerPolicy) IsPIIRestricted(filePath string) bool {
	for _, pattern := range p.PIIRestricted {
		if pathMatches(filePath, pattern) {
			return true
		}
	}
	return false
}

func pathMatches(filePath, pattern string) bool {
	if pattern == "" {
		return false
	}
	normPath := strings.ReplaceAll(filePath, "\\", "/")
	if !strings.HasPrefix(normPath, "/") {
		normPath = "/" + normPath
	}
	return strings.Contains(normPath, pattern)
}

// normalizePathPattern turns "analysis" or "models/analysis" into "/analysis/" for segment matching.
func normalizePathPattern(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	p = strings.ReplaceAll(p, "\\", "/")
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if !strings.HasSuffix(p, "/") {
		p = p + "/"
	}
	return p
}
