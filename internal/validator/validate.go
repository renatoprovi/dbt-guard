package validator

import (
	"github.com/renatocruz/dbt-guard/internal/parser"
)

// Violation represents an analysis-layer model that descends from PII without masking.
type Violation struct {
	ModelID     string   // model unique_id
	LineagePath []string // path from the model to a PII source/node
}

// RunValidate loads the manifest, finds models under analysis/, and returns violations:
// models that descend from PII and are not explicitly masked (meta.masked).
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
