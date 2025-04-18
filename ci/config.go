package main

import (
	"bytes"
	"context"
	"dagger/ci/internal/dagger"
	"dagger/ci/tmpls"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/goccy/go-yaml"
)

// ConfString holds a json of [Conf].
type ConfString = string

func ReadConfString(s string) (Conf, error) {
	var c Conf
	if err := json.Unmarshal([]byte(s), &c); err != nil {
		return Conf{}, err
	}
	return c, nil
}

type Conf struct {
	Jobs []Job `json:"jobs"` // Wow
}

// // Returns the json schema that the config file follows.
// func (m *Ci) ConfigJsonSchema() string {
// 	r := &jsonschema.Reflector{
// 		KeyNamer:       strcase.KebabCase,
// 		ExpandedStruct: true,
// 	}

// 	json, _ := r.Reflect(m.Conf).MarshalJSON()
// 	return string(json)
// }

func (Ci) parseConfFile(ctx context.Context, cfgFile *dagger.File) (ConfString, error) {

	if _, err := cfgFile.Sync(ctx); err != nil {
		return "", fmt.Errorf("Config file was not accessible: %w", err)
	}

	cfgContents, err := cfgFile.Contents(ctx)
	if err != nil {
		return "", fmt.Errorf("Config file contents could not be read: %w", err)
	}
	cfgFileName, err := cfgFile.Name(ctx)
	if err != nil {
		return "", fmt.Errorf("Config file name could not be retrieved: %w", err)
	}

	var cs bytes.Buffer
	// Interpret templates.
	tmpl, err := template.
		New(cfgFileName).
		Funcs(tmpls.TmplFuncs()).
		Parse(cfgContents)
	if err != nil {
		return "", err
	}
	if err := tmpl.Execute(&cs, nil); err != nil {
		return "", err
	}

	// Unmarshal config.
	var c Conf
	if err := yaml.UnmarshalWithOptions(cs.Bytes(), &c, yaml.AllowDuplicateMapKey()); err != nil {
		return "", fmt.Errorf("Couldnt unmarshal config file %s: %w", cfgFileName, err)
	}

	// Serialize config to JSON.
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("Couldnt marshal config file %s: %w", cfgFileName, err)
	}

	return string(jsonBytes), nil

}

func (b *Builder) PrintConf() string {
	return b.Conf
}
