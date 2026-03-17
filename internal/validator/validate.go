package validator

import (
	"github.com/renatocruz/dbt-guard/internal/parser"
)

// Violation representa um modelo na camada analysis que descende de PII sem mascaramento.
type Violation struct {
	ModelID     string   // unique_id do modelo
	LineagePath []string // caminho do modelo até uma source/nó PII
}

// RunValidate carrega o manifest, identifica modelos em analysis/, e retorna violações:
// modelos que descendem de PII e não estão marcados como mascarados (meta.masked).
func RunValidate(manifestPath string) ([]Violation, error) {
	m, err := parser.LoadManifest(manifestPath)
	if err != nil {
		return nil, err
	}
	var violations []Violation
	for _, nodeID := range m.AnalysisNodeIDs() {
		node := m.Nodes[nodeID]
		if node == nil {
			continue
		}
		if !parser.IsSensitive(nodeID, m) {
			continue
		}
		if parser.IsNodeMasked(node) {
			continue
		}
		path := parser.LineagePathToPII(nodeID, m)
		if len(path) > 0 {
			violations = append(violations, Violation{ModelID: nodeID, LineagePath: path})
		}
	}
	return violations, nil
}
