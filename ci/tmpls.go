package main

import (
	"text/template"
	"time"
)

// TagTmplFuncs returns a template.FuncMap with functions used in tag templates.
func TagTmplFuncs() template.FuncMap {
	return template.FuncMap{"now": func() string {
		return time.Now().UTC().Format("20060102")
	},
	}
}
