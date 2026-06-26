package parser

import "fmt"

// PIIColumn describes a column marked as PII in a file/table.
type PIIColumn struct {
	FilePath string
	Source   string
	Table    string
	Column   string
}

// CollectPIIColumns walks a SourceFile and returns every column with security_tag pii.
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

// PrintPIIColumns finds all sources.yml files under root, parses them, and prints each PII column name.
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
