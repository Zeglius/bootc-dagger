package main

import (
	"context"
	"dagger/ci/internal/dagger"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"text/template"

	"github.com/goccy/go-yaml"
	"github.com/invopop/jsonschema"
	"github.com/stoewer/go-strcase"
)

type Conf struct {
	Jobs []Job `json:"jobs"` // Wow
}

// Load configuration file
func (m *Ci) WithConfig(
	ctx context.Context,
	// +defaultPath="."
	dir *dagger.Directory,
	// +defaultPath="bootc-ci.yaml"
	cfgFile *dagger.File,
) (*Ci, error) {

	m.BuildContext = dir
	c, err := m.parseConfFile(ctx, cfgFile)
	if err != nil {
		return nil, err
	}
	m.Conf = c

	return m, nil
}

// Returns the json schema that the config file follows.
func (m *Ci) ConfigJsonSchema() string {
	r := &jsonschema.Reflector{
		KeyNamer:       strcase.KebabCase,
		ExpandedStruct: true,
	}

	json, _ := r.Reflect(m.Conf).MarshalJSON()
	return string(json)
}

func (*Ci) parseConfFile(ctx context.Context, cfgFile *dagger.File) (*Conf, error) {
	if _, err := cfgFile.Sync(ctx); err != nil {
		return nil, fmt.Errorf("Config file was not accessible: %w", err)
	}

	cfgContents, err := cfgFile.Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("Config file contents could not be read: %w", err)
	}
	cfgFileName, err := cfgFile.Name(ctx)
	if err != nil {
		return nil, fmt.Errorf("Config file name could not be retrieved: %w", err)
	}

	result := &Conf{}
	switch path.Ext(cfgFileName) {
	case ".yml":
		fallthrough
	case ".yaml":
		if err := yaml.UnmarshalWithOptions([]byte(cfgContents), result, yaml.AllowDuplicateMapKey()); err != nil {
			return nil, fmt.Errorf("Failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal([]byte(cfgContents), result); err != nil {
			return nil, fmt.Errorf("Failed to parse JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("Unsupported config file format: %s", cfgFileName)
	}

	// Parse tags. These can have golang templates
	tmpl := template.New("tags").
		Funcs(TagTmplFuncs())
	for i, j := range result.Jobs { // For each job
		for i2, t := range j.OutputTags { // For each tag
			var s strings.Builder
			if tmpl, err := tmpl.Parse(t); err != nil {
				return nil, err
			} else {
				tmpl.Execute(&s, j)
			}
			// Replace the text with the parsed template
			result.Jobs[i].OutputTags[i2] = s.String()

		}
	}

	return result, nil
}

func (m *Ci) PrintConf() (string, error) {
	b, err := json.Marshal(m.Conf)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
