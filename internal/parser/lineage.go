package parser

// IsSensitive reports whether nodeID descends from any node or source with security_tag: pii.
// It walks the graph depth-first (DFS) from nodeID following depends_on (parents).
// Returns true if any ancestor is a source with a PII column or a node with meta.security_tag: pii.
// Uses a per-nodeID cache to avoid cycles and re-evaluation (each node is evaluated at most once).
func IsSensitive(nodeID string, m *Manifest) bool {
	if m == nil {
		return false
	}
	cache := make(map[string]bool)
	return isSensitiveDFS(nodeID, m, cache)
}

func isSensitiveDFS(nodeID string, m *Manifest, cache map[string]bool) bool {
	if v, ok := cache[nodeID]; ok {
		return v
	}
	if src, ok := m.Sources[nodeID]; ok {
		v := src.HasPIIColumn()
		cache[nodeID] = v
		return v
	}
	node, ok := m.Nodes[nodeID]
	if !ok {
		cache[nodeID] = false
		return false
	}
	if hasPIITag(node.Meta) || (node.Config != nil && hasPIITag(node.Config.Meta)) {
		cache[nodeID] = true
		return true
	}
	if node.DependsOn == nil {
		cache[nodeID] = false
		return false
	}
	for _, parentID := range node.DependsOn.Nodes {
		if isSensitiveDFS(parentID, m, cache) {
			cache[nodeID] = true
			return true
		}
	}
	cache[nodeID] = false
	return false
}

// LineagePathToPII returns a path from nodeID to a PII source/node (first match in DFS).
// The returned slice is [nodeID, ... parent ..., piiNodeID]. Empty if the node does not descend from PII.
func LineagePathToPII(nodeID string, m *Manifest) []string {
	if m == nil {
		return nil
	}
	visited := make(map[string]bool)
	return lineagePathDFS(nodeID, m, visited, nil)
}

func lineagePathDFS(nodeID string, m *Manifest, visited map[string]bool, path []string) []string {
	if visited[nodeID] {
		return nil
	}
	visited[nodeID] = true
	path = append(path, nodeID)
	if src, ok := m.Sources[nodeID]; ok {
		if src.HasPIIColumn() {
			return path
		}
		return nil
	}
	node, ok := m.Nodes[nodeID]
	if !ok {
		return nil
	}
	if hasPIITag(node.Meta) || (node.Config != nil && hasPIITag(node.Config.Meta)) {
		return path
	}
	if node.DependsOn == nil {
		return nil
	}
	for _, parentID := range node.DependsOn.Nodes {
		pathCopy := append([]string(nil), path...)
		if p := lineagePathDFS(parentID, m, visited, pathCopy); len(p) > 0 {
			return p
		}
	}
	return nil
}
