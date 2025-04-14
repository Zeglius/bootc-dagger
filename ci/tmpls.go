package main

import (
	"encoding/json"
	"text/template"
	"time"
)

// TmplFuncs returns a template.FuncMap with functions used in tag templates.
func TmplFuncs() template.FuncMap {
	return template.FuncMap{

		"nowTag": func() string {
			return time.Now().UTC().Format("20060102")
		},

		"json": func(a any) string {
			b, _ := json.Marshal(a)

			return string(b)
		},
	}
}
