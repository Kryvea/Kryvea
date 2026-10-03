package model

const (
	SourceGeneric = "generic"
	SourceNessus  = "nessus"
	SourceBurp    = "burp"
)

type Category struct {
	Model
	Identifier         string            `json:"identifier"`
	Name               string            `json:"name"`
	Subcategory        string            `json:"subcategory,omitempty"`
	GenericDescription map[string]string `json:"generic_description,omitempty"`
	GenericRemediation map[string]string `json:"generic_remediation,omitempty"`
	LanguagesOrder     []string          `json:"languages_order,omitempty"`
	References         []string          `json:"references"`
	Source             string            `json:"source"`
}
