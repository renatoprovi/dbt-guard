package parser

import "fmt"

// PIIColumn descreve uma coluna marcada como PII em um arquivo/tabela.
type PIIColumn struct {
	FilePath string
	Source   string
	Table    string
	Column   string
}

// CollectPIIColumns percorre um SourceFile e retorna todas as colunas com security_tag pii.
func CollectPIIColumns(filePath string, sf *SourceFile) []PIIColumn {
	var out []PIIColumn
	for _, src := range sf.Sources {
		for _, tbl := range src.Tables {
			for _, col := range tbl.Columns {
				if col.IsPII() {
					out = append(out, PIIColumn{
						FilePath: filePath,
						Source:   src.Name,
						Table:    tbl.Name,
						Column:   col.Name,
					})
				}
			}
		}
	}
	return out
}

// PrintPIIColumns encontra todos os sources.yml em root, faz parse e imprime o nome de cada coluna PII.
func PrintPIIColumns(root string) error {
	paths, err := FindSourceFiles(root)
	if err != nil {
		return err
	}
	for _, path := range paths {
		sf, err := ParseSourceFilePath(path)
		if err != nil {
			return err
		}
		for _, pii := range CollectPIIColumns(path, sf) {
			fmt.Println(pii.Column)
		}
	}
	return nil
}
