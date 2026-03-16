package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Manifest representa o manifest.json do dbt (v10+).
// Apenas nodes e sources são mapeados; o decoder ignora o resto (otimização de memória).
type Manifest struct {
	Metadata json.RawMessage          `json:"metadata,omitempty"`
	Nodes    map[string]*ManifestNode `json:"nodes"`
	Sources  map[string]*SourceDef    `json:"sources"`
}

// ManifestNode representa um nó do grafo (model, analysis, seed, etc.).
type ManifestNode struct {
	UniqueID         string      `json:"unique_id"`
	ResourceType     string      `json:"resource_type"`
	DependsOn        *DependsOn  `json:"depends_on,omitempty"`
	Meta             MetaMap     `json:"meta,omitempty"`
	Config           *ConfigMeta `json:"config,omitempty"`
	OriginalFilePath string      `json:"original_file_path,omitempty"`
	Name             string      `json:"name,omitempty"`
	Fqn              []string    `json:"fqn,omitempty"`
	// Campos não usados na linhagem podem ser ignorados; o decoder preenche só o que existe.
}

// DependsOn contém as dependências do nó (parents no grafo).
type DependsOn struct {
	Nodes  []string `json:"nodes"`
	Macros []string `json:"macros"`
}

// MetaMap armazena meta (ex.: security_tag) como chave/valor.
type MetaMap map[string]interface{}

// ConfigMeta é o bloco config do dbt (v1.10+); meta pode estar aqui.
type ConfigMeta struct {
	Meta MetaMap `json:"meta,omitempty"`
}

// SourceDef representa uma source no manifest (fonte de dados declarada).
type SourceDef struct {
	UniqueID         string                `json:"unique_id"`
	SourceName       string                `json:"source_name"`
	Name             string                `json:"name"`
	Columns          map[string]ColumnInfo `json:"columns,omitempty"`
	Meta             MetaMap               `json:"meta,omitempty"`
	OriginalFilePath string                `json:"original_file_path,omitempty"`
}

// ColumnInfo descreve uma coluna (ex.: em uma source); meta pode conter security_tag.
type ColumnInfo struct {
	Meta   MetaMap     `json:"meta,omitempty"`
	Config *ConfigMeta `json:"config,omitempty"`
}

// LoadManifest lê o arquivo em path e faz unmarshal para Manifest.
// Usa apenas os campos mapeados nas structs; o resto não é alocado (decoder ignora).
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

// NodeIDsWithPII retorna os unique_id dos nós (models, etc.) que têm meta.security_tag == "pii".
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

// SourceIDsWithPII retorna os unique_id das sources que possuem alguma coluna com security_tag == "pii".
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

// HasPIIColumn retorna true se alguma coluna da source tiver meta.security_tag == "pii".
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

// IsNodeMasked retorna true se o nó está explicitamente marcado como mascarado
// (meta.masked ou config.meta.masked == true). Usado pelo validador da camada analysis.
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

// AnalysisNodeIDs retorna os unique_id dos nós que estão na pasta analysis
// (original_file_path contém "/analysis/").
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

// PrintManifestPII carrega o manifest em path e imprime os unique_id de nós e sources com tag PII.
// Usado pelo comando "dbt-guard manifest <path>".
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

// PrintSensitiveNodes carrega o manifest e imprime os unique_id de todos os nós e sources
// que são sensíveis (descendem de PII ou são PII). Usa IsSensitive (DFS).
// Usado pelo comando "dbt-guard sensitive <path>".
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
