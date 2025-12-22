package config

import "gopkg.in/yaml.v3"

type AppConfig struct {
	Title     string     `yaml:"title" json:"title"`
	Providers []string   `yaml:"providers" json:"providers"`
	Menu      []MenuItem `yaml:"menu" json:"menu"`
	Pages     []Page     `yaml:"pages" json:"pages"`
}

type MenuItem struct {
	Title    string     `yaml:"title" json:"title"`
	Page     string     `yaml:"page" json:"page,omitempty"`
	Href     string     `yaml:"href" json:"href,omitempty"`
	Children []MenuItem `yaml:"children" json:"children,omitempty"`
}

type Page struct {
	Slug    string   `yaml:"slug" json:"slug"`
	Title   string   `yaml:"title" json:"title"`
	Widgets []Widget `yaml:"widgets" json:"widgets"`
}

type Widget struct {
	ID       string       `yaml:"id" json:"id"`
	Title    string       `yaml:"title" json:"title"`
	Type     string       `yaml:"type" json:"type"`
	Provider ProviderSpec `yaml:"provider" json:"-"` // exclude from /api/config response for security reasons
	Table    *TableSpec   `yaml:"table" json:"table,omitempty"`
}

type ProviderSpec struct {
	Name string   `yaml:"name" json:"name"`
	SQL  *SQLSpec `yaml:"sql" json:"sql,omitempty"`
}

type SQLSpec struct {
	Query      string            `yaml:"query" json:"query"`
	Bindings   map[string]string `yaml:"bindings" json:"bindings"`
	Pagination *PaginationSpec   `yaml:"pagination" json:"pagination,omitempty"`
}

type PaginationSpec struct {
	Column string `yaml:"column" json:"column"`
	Order  string `yaml:"order" json:"order"`
}

type TableSpec struct {
	Columns []ColumnSpec `yaml:"columns" json:"columns"`
	Filters []FilterSpec `yaml:"filters" json:"filters,omitempty"`
}

type ColumnSpec struct {
	ID     string        `yaml:"id" json:"id"`
	Title  string        `yaml:"title" json:"title,omitempty"`
	Render *ColumnRender `yaml:"render" json:"render,omitempty"`
}

type ColumnRender struct {
	Type     string `yaml:"type" json:"type"`
	Text     string `yaml:"text" json:"text,omitempty"`
	URL      string `yaml:"url" json:"url,omitempty"`
	External bool   `yaml:"external" json:"external,omitempty"`
}

func (c *ColumnSpec) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		c.ID = n.Value
		c.Title = n.Value
		return nil
	}

	type raw ColumnSpec
	var parsed raw
	if err := n.Decode(&parsed); err != nil {
		return err
	}

	if parsed.Title == "" {
		parsed.Title = parsed.ID
	}
	*c = ColumnSpec(parsed)
	return nil
}

type FilterSpec struct {
	ID        string        `yaml:"id" json:"id"`
	Title     string        `yaml:"title" json:"title"`
	Type      string        `yaml:"type" json:"type"`
	Column    string        `yaml:"column" json:"column"`
	Mode      string        `yaml:"mode" json:"mode,omitempty"`
	Operators []string      `yaml:"operators" json:"operators,omitempty"`
	Values    []ValueOption `yaml:"values" json:"values,omitempty"`
}

type ValueOption struct {
	Value string `yaml:"value" json:"value"`
	Label string `yaml:"label" json:"label"`
}
