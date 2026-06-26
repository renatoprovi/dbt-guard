package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Manifest represents a dbt manifest.json (v10+).
// Only nodes and sources are mapped; the decoder ignores the rest (memory optimization).
type Manifest struct {
	Metadata json.RawMessage          `json:"metadata,omitempty"`
	Nodes    map[string]*ManifestNode `json:"nodes"`
	Sources  map[string]*SourceDef    `json:"sources"`
}

// ManifestNode represents a graph node (model, analysis, seed, etc.).
type ManifestNode struct {
	UniqueID         string      `json:"unique_id"`
	ResourceType     string      `json:"resource_type"`
	DependsOn        *DependsOn  `json:"depends_on,omitempty"`
	Meta             MetaMap     `json:"meta,omitempty"`
	Config           *ConfigMeta `json:"config,omitempty"`
	OriginalFilePath string      `json:"original_file_path,omitempty"`
	Name             string      `json:"name,omitempty"`
	Fqn              []string    `json:"fqn,omitempty"`
	// Fields unused for lineage can be omitted; the decoder fills only what exists.
}

// DependsOn holds node dependencies (parents in the graph).
type DependsOn struct {
	Nodes  []string `json:"nodes"`
	Macros []string `json:"macros"`
}

// MetaMap stores meta fields (e.g. security_tag) as key/value pairs.
type MetaMap map[string]interface{}

// ConfigMeta is the dbt config block (v1.10+); meta may live here.
type ConfigMeta struct {
	Meta MetaMap `json:"meta,omitempty"`
}

// SourceDef represents a source in the manifest (declared data source).
type SourceDef struct {
	UniqueID         string                `json:"unique_id"`
	SourceName       string                `json:"source_name"`
	Name             string                `json:"name"`
	Columns          map[string]ColumnInfo `json:"columns,omitempty"`
	Meta             MetaMap               `json:"meta,omitempty"`
	OriginalFilePath string                `json:"original_file_path,omitempty"`
}

// ColumnInfo describes a column (e.g. on a source); meta may contain security_tag.
type ColumnInfo struct {
	Meta   MetaMap     `json:"meta,omitempty"`
	Config *ConfigMeta `json:"config,omitempty"`
}

// LoadManifest reads the file at path and unmarshals it into Manifest.
// Only mapped struct fields are allocated; the decoder ignores the rest.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.Nodes == nil {
		m.Nodes = make(map[string]*ManifestNode)
	}
	if m.Sources == nil {
		m.Sources = make(map[string]*SourceDef)
	}
	return &m, nil
}

// NodeIDsWithPII returns unique_id values for nodes (models, etc.) with meta.security_tag == "pii".
func (m *Manifest) NodeIDsWithPII() []string {
	var out []string
	for id, n := range m.Nodes {
		if n == nil {
			continue
		}
		if hasPIITag(n.Meta) || (n.Config != nil && hasPIITag(n.Config.Meta)) {
			out = append(out, id)
		}
	}
	return out
}

// SourceIDsWithPII returns unique_id values for sources that have at least one column with security_tag == "pii".
func (m *Manifest) SourceIDsWithPII() []string {
	var out []string
	for id, s := range m.Sources {
		if s == nil {
			continue
		}
		if s.HasPIIColumn() {
			out = append(out, id)
		}
	}
	return out
}

// HasPIIColumn reports whether any column on the source has meta.security_tag == "pii".
func (s *SourceDef) HasPIIColumn() bool {
	for _, c := range s.Columns {
		if hasPIITag(c.Meta) || (c.Config != nil && hasPIITag(c.Config.Meta)) {
			return true
		}
	}
	return false
}

func hasPIITag(meta MetaMap) bool {
	if meta == nil {
		return false
	}
	v, ok := meta["security_tag"]
	if !ok {
		return false
	}
	tag, _ := v.(string)
	return tag == "pii"
}

// IsNodeMasked reports whether the node is explicitly marked as masked
// (meta.masked or config.meta.masked == true). Used by the analysis-layer validator.
func IsNodeMasked(n *ManifestNode) bool {
	if n == nil {
		return false
	}
	if isMaskedMeta(n.Meta) {
		return true
	}
	if n.Config != nil && isMaskedMeta(n.Config.Meta) {
		return true
	}
	return false
}

func isMaskedMeta(meta MetaMap) bool {
	if meta == nil {
		return false
	}
	v, ok := meta["masked"]
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}

// AnalysisNodeIDs returns unique_id values for nodes under the analysis folder
// (original_file_path contains "/analysis/").
func (m *Manifest) AnalysisNodeIDs() []string {
	var out []string
	for id, n := range m.Nodes {
		if n == nil {
			continue
		}
		if strings.Contains(n.OriginalFilePath, "/analysis/") {
			out = append(out, id)
		}
	}
	return out
}

// PrintManifestPII loads the manifest at path and prints unique_id values for nodes and sources tagged as PII.
// Used by the "dbt-guard manifest <path>" command.
func PrintManifestPII(path string) error {
	m, err := LoadManifest(path)
	if err != nil {
		return err
	}
	for _, id := range m.NodeIDsWithPII() {
		fmt.Println(id)
	}
	for _, id := range m.SourceIDsWithPII() {
		fmt.Println(id)
	}
	return nil
}

// PrintSensitiveNodes loads the manifest and prints unique_id values for all nodes and sources
// that are sensitive (declare PII or descend from PII). Uses IsSensitive (DFS).
// Used by the "dbt-guard sensitive <path>" command.
func PrintSensitiveNodes(path string) error {
	m, err := LoadManifest(path)
	if err != nil {
		return err
	}
	for id := range m.Nodes {
		if IsSensitive(id, m) {
			fmt.Println(id)
		}
	}
	for id := range m.Sources {
		if IsSensitive(id, m) {
			fmt.Println(id)
		}
	}
	return nil
}
