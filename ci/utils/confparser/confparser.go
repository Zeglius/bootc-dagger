// Handles template configuration parsing for builder CI.
//
// It exposes [Parse] which takes a template string and options for customizing
// the parsing behavior.
package confparser

import (
	"bytes"
	"dagger/ci/config"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/goccy/go-yaml"
)

type ParseOpts struct {
	TmplName  string
	TmplFuncs []template.FuncMap
	Data      map[string]any
}

// Parse takes a template string and returns a [config.ConfString].
// The template string is parsed and executed with the given options.
// ParseOpts provides the template functions and dagger client to be used.
func Parse(
	tmpltext string,
	opts ParseOpts,
) (config.ConfString, error) {

	// Set a template name by default
	if opts.TmplName == "" {
		opts.TmplName = "<config_file>"
	}

	t := template.New(opts.TmplName)

	// Feed template funcs
	for _, funcs := range opts.TmplFuncs {
		t = t.Funcs(funcs)
	}

	// Parse and execute template
	t, err := t.Parse(tmpltext)
	if err != nil {
		return "", fmt.Errorf("Couldnt parse config template: %w", err)
	}
	var cs bytes.Buffer
	t.Execute(&cs, opts.Data)

	// Unmarshal config
	var conf config.Conf
	if err := yaml.UnmarshalWithOptions(cs.Bytes(), &conf, yaml.AllowDuplicateMapKey()); err != nil {
		return "", fmt.Errorf("Couldnt unmarshal config file %s: %w", opts.TmplName, err)
	}

	// Lowercase registry
	{
		old_conf := conf
		for ji, job := range old_conf.Jobs {
			conf.Jobs[ji].OutputName = strings.ToLower(job.OutputName)
		}
	}

	// Serialize config to json
	jsonBytes, err := json.Marshal(conf)
	if err != nil {
		return "", fmt.Errorf("Couldnt marshal config: %w", err)
	}

	return config.ConfString(jsonBytes), nil

}
