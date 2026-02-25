package models

import "encoding/json"

type FormSchema struct {
	ID    string `json:"formId"`
	Title string `json:"formTitle"`
	Nodes []Node  `json:"nodes"`
}

type FormSchemaDB struct {
	Title string `json:"formTitle"`
	Nodes []Node  `json:"nodes"`
}

type Node struct {
	Type string `json:"nodeType"`
	ID string `json:"nodeId"`
	Nodes []Node `json:"nodes"`
	Props []Prop `json:"props"`
}

type Prop struct {
	Type string `json:"propType"`
	Value any `json:"value"`
	Editable bool `json:"editable"`
}

type DateRange struct {
	Max string `json:"max"`
	Min string `json:"min"`
}

func (p *Prop) UnmarshalJSON(data []byte) error {
	type Alias Prop
	aux := &struct {
		Value json.RawMessage `json:"value"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch p.Type {

	// string props
	case "label", "placeholder":
		var v string
		if err := json.Unmarshal(aux.Value, &v); err != nil {
			return err
		}
		p.Value = v

	// numeric props
	case "maxlengthchar", "maxlengthword":
		var v float64
		if err := json.Unmarshal(aux.Value, &v); err != nil {
			return err
		}
		p.Value = v

	// boolean props
	case "required", "email", "patternphone", "patternnumber", "optionother", "allowtoggle":
		var v bool
		if err := json.Unmarshal(aux.Value, &v); err != nil {
			return err
		}
		p.Value = v

	// array props
	case "options":
		var v []string
		if err := json.Unmarshal(aux.Value, &v); err != nil {
			return err
		}
		p.Value = v

	// custom struct
	case "daterange":
		var v DateRange
		if err := json.Unmarshal(aux.Value, &v); err != nil {
			return err
		}
		p.Value = v

	// fallback
	default:
		p.Value = aux.Value
	}

	return nil
}