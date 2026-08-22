package parser

import (
	"github.com/renatoprovi/dbt-guard/internal/config"
)

// RestrictedNodeIDs returns unique_id values for models in restricted layers that are not in allowed layers.
// Allowed layers take precedence when a path matches both lists.
func (m *Manifest) RestrictedNodeIDs(policy config.LayerPolicy) []string {
	var out []string
	for id, n := range m.Nodes {
		if n == nil {
			continue
		}
		path := n.OriginalFilePath
		if policy.IsPIIAllowed(path) {
			continue
		}
		if policy.IsPIIRestricted(path) {
			out = append(out, id)
		}
	}
	return out
}
