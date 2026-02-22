package models

type FormSchema struct {
	ID    string `json:"formId"`
	Title string `json:"formTitle"`
	Nodes []any  `json:"nodes"`
}

type FormSchemaDB struct {
	Title string `json:"formTitle"`
	Nodes []any  `json:"nodes"`
}