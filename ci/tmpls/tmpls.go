package tmpls

import (
	"encoding/json"
	"maps"
	"strings"
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

// readKeyVal reads key-value pairs from a slice of strings and returns the value for the given key.
func readKeyVal(k string, s []string) string {
	vals := maps.Collect(func(yield func(string, string) bool) {
		for _, v := range s {
			k, v1, _ := strings.Cut(v, "=")
			if !yield(k, v1) {
				return
			}
		}
	})

	v, _ := vals[k]
	return v
}

// TmplFuncs returns a template.FuncMap with functions used in tag templates.
func TmplFuncs() template.FuncMap {
	return template.FuncMap{
		"nowTag":     nowTag,
		"json":       jsonMarshal,
		"readKeyVal": readKeyVal,
	}
}
