package parser

// IsSensitive indica se o nó nodeID descende de algum nó ou source com security_tag: pii.
// Percorre o grafo em profundidade (DFS) a partir de nodeID seguindo depends_on (parents).
// Se algum ancestral for uma source com coluna PII ou um node com meta.security_tag: pii,
// retorna true. Usa cache por nodeID para evitar ciclos e reavaliação (cada nó é avaliado no máximo uma vez).
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

// LineagePathToPII retorna um caminho de nodeID até uma source/nó PII (primeiro encontrado em DFS).
// O slice retornado é [nodeID, ... parent ..., piiNodeID]. Vazio se o nó não descende de PII.
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
