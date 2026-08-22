package parser

import (
	"os"

	"gopkg.in/yaml.v3"
)

// SourceFile represents the contents of a dbt sources YAML file (e.g. sources.yml).
type SourceFile struct {
	Sources []Source `yaml:"sources"`
}

// Source represents a dbt source (origin schema/tables).
type Source struct {
	Name   string  `yaml:"name"`
	Tables []Table `yaml:"tables"`
}

// Table represents a table within a source.
type Table struct {
	Name    string   `yaml:"name"`
	Columns []Column `yaml:"columns"`
}

// Column represents a column; meta holds security_tag (e.g. pii).
type Column struct {
	Name   string        `yaml:"name"`
	Meta   *ColumnMeta   `yaml:"meta,omitempty"`
	Config *ColumnConfig `yaml:"config,omitempty"`
}

// ColumnMeta holds column metadata, including security_tag.
type ColumnMeta struct {
	SecurityTag string `yaml:"security_tag"`
}

// ColumnConfig is the dbt config block (v1.10+); meta may live here.
type ColumnConfig struct {
	Meta *ColumnMeta `yaml:"meta,omitempty"`
}

// ParseSourceFile parses YAML content from a sources file.
func ParseSourceFile(content []byte) (*SourceFile, error) {
	var f SourceFile
	if err := yaml.Unmarshal(content, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// ParseSourceFilePath reads and parses the file at the given path.
func ParseSourceFilePath(path string) (*SourceFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseSourceFile(content)
}

// SecurityTag returns the column security_tag (meta or config.meta).
// Returns an empty string when no tag is set.
func (c *Column) SecurityTag() string {
	if c.Meta != nil && c.Meta.SecurityTag != "" {
		return c.Meta.SecurityTag
	}
	if c.Config != nil && c.Config.Meta != nil && c.Config.Meta.SecurityTag != "" {
		return c.Config.Meta.SecurityTag
	}
	return ""
}

// IsPII reports whether the column has security_tag == "pii".
func (c *Column) IsPII() bool {
	return c.SecurityTag() == "pii"
}
