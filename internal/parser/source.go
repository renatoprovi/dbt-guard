package parser

import (
	"os"

	"gopkg.in/yaml.v3"
)

// SourceFile representa o conteúdo de um arquivo YAML de sources do dbt (ex.: sources.yml).
type SourceFile struct {
	Version int      `yaml:"version,omitempty"`
	Sources []Source `yaml:"sources"`
}

// Source representa um source do dbt (schema/tabelas de origem).
type Source struct {
	Name   string  `yaml:"name"`
	Schema string  `yaml:"schema,omitempty"`
	Tables []Table `yaml:"tables"`
}

// Table representa uma tabela dentro de um source.
type Table struct {
	Name    string   `yaml:"name"`
	Columns []Column `yaml:"columns"`
}

// Column representa uma coluna; meta guarda security_tag (ex.: pii).
type Column struct {
	Name   string        `yaml:"name"`
	Meta   *ColumnMeta   `yaml:"meta,omitempty"`
	Config *ColumnConfig `yaml:"config,omitempty"`
}

// ColumnMeta contém metadados da coluna, incluindo security_tag.
type ColumnMeta struct {
	SecurityTag string `yaml:"security_tag"`
}

// ColumnConfig é o bloco config do dbt (v1.10+); meta pode estar aqui.
type ColumnConfig struct {
	Meta *ColumnMeta `yaml:"meta,omitempty"`
}

// ParseSourceFile lê e faz parse do conteúdo YAML de um arquivo de sources.
func ParseSourceFile(content []byte) (*SourceFile, error) {
	var f SourceFile
	if err := yaml.Unmarshal(content, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// ParseSourceFilePath lê o arquivo no caminho dado e faz parse.
func ParseSourceFilePath(path string) (*SourceFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseSourceFile(content)
}

// SecurityTag retorna a security_tag da coluna (meta ou config.meta).
// Retorna string vazia se não houver tag.
func (c *Column) SecurityTag() string {
	if c.Meta != nil && c.Meta.SecurityTag != "" {
		return c.Meta.SecurityTag
	}
	if c.Config != nil && c.Config.Meta != nil && c.Config.Meta.SecurityTag != "" {
		return c.Config.Meta.SecurityTag
	}
	return ""
}

// IsPII retorna true se a coluna tiver security_tag == "pii".
func (c *Column) IsPII() bool {
	return c.SecurityTag() == "pii"
}
