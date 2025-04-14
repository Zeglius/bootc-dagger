package tmpls

import (
	"encoding/json"
	"text/template"
	"time"
)

func nowTag() string {
	return time.Now().UTC().Format("20060102")
}

func jsonMarshal(a any) string {
	b, _ := json.Marshal(a)

	return string(b)
}

// TmplFuncs returns a template.FuncMap with functions used in tag templates.
func TmplFuncs() template.FuncMap {
	return template.FuncMap{
		"nowTag": nowTag,
		"json":   jsonMarshal,
	}
}
