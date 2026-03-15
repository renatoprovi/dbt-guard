package parser

import (
	"io/fs"
	"path/filepath"
)

const sourcesFileName = "sources.yml"

// FindSourceFiles percorre root recursivamente e retorna os caminhos de todos os arquivos sources.yml.
func FindSourceFiles(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) == sourcesFileName {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}
