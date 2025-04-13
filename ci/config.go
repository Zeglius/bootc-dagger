package main

import (
	"context"
	"dagger/ci/internal/dagger"
	"encoding/json"
	"fmt"
	"path"

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
	r.AddGoComments("dagger/ci", "./", jsonschema.WithFullComment())

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
		yaml.Unmarshal([]byte(cfgContents), result)
	case ".json":
		json.Unmarshal([]byte(cfgContents), result)
	default:
		return nil, fmt.Errorf("Unsupported config file format: %s", cfgFileName)
	}

	return result, nil
}
