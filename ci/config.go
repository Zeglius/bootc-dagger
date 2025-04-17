package main

import (
	"context"
	"dagger/ci/internal/dagger"
	"dagger/ci/tmpls"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"text/template"

	"github.com/goccy/go-yaml"
)

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

// We cant use map[string]any (or any map) in Job struct, as dagger codegen will
// refuse to work with these if used in public facing types.
//
// Instead, we will use an anonymous struct that uses map, and a anonymous function
// that replace these with [Pair].
type JobParseable struct {
	Containerfile string            `json:"containerfile"`
	BuildArgs     map[string]string `json:"build-args,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	OutputName    string            `json:"output-name,omitempty"`
	OutputTags    []string          `json:"output-tags,omitempty"`
}

type ConfParseable struct {
	Jobs []JobParseable `json:"jobs"`
}

// Convert a [JobParseable] to a [Job]
func toJob(j JobParseable) Job {
	res := Job{
		Containerfile: j.Containerfile,
		BuildArgs:     make([]Pair, 0, len(j.BuildArgs)),
		Annotations:   make([]Pair, 0, len(j.Annotations)),
		Labels:        make([]Pair, 0, len(j.Labels)),
		OutputName:    j.OutputName,
		OutputTags:    j.OutputTags,
	}
	for k, v := range j.BuildArgs {
		res.BuildArgs = append(res.BuildArgs, Pair{Key: k, Value: v})
	}
	for k, v := range j.Annotations {
		res.Annotations = append(res.Annotations, Pair{Key: k, Value: v})
	}
	for k, v := range j.Labels {
		res.Labels = append(res.Labels, Pair{Key: k, Value: v})
	}

	return res
}

func (Ci) parseConfFile(ctx context.Context, cfgFile *dagger.File) (*Conf, error) {

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

	c := &ConfParseable{}
	switch path.Ext(cfgFileName) {
	case ".yml":
		fallthrough
	case ".yaml":
		if err := yaml.UnmarshalWithOptions([]byte(cfgContents), c, yaml.AllowDuplicateMapKey()); err != nil {
			return nil, fmt.Errorf("Failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal([]byte(cfgContents), c); err != nil {
			return nil, fmt.Errorf("Failed to parse JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("Unsupported config file format: %s", cfgFileName)
	}

	result := &Conf{}
	for _, jp := range c.Jobs {
		result.Jobs = append(result.Jobs, toJob(jp))
	}

	// Parse templates. These can have golang templates
	tmpl := template.New("templates").
		Funcs(tmpls.TmplFuncs())
	for i, j := range result.Jobs { // For each job

		// Process output tags
		for i2, t := range j.OutputTags {
			var s strings.Builder
			if tmpl, err := tmpl.Parse(t); err != nil {
				return nil, err
			} else {
				tmpl.Execute(&s, j)
			}
			result.Jobs[i].OutputTags[i2] = s.String()
		}

		// Process build args
		for k, ba := range j.BuildArgs {
			var s strings.Builder
			if tmpl, err := tmpl.Parse(ba.Value); err != nil {
				return nil, err
			} else {
				tmpl.Execute(&s, j)
			}
			result.Jobs[i].BuildArgs[k].Value = s.String()
		}

		// Process annotations
		for k, annot := range j.Annotations {
			var s strings.Builder
			if tmpl, err := tmpl.Parse(annot.Value); err != nil {
				return nil, err
			} else {
				tmpl.Execute(&s, j)
			}
			result.Jobs[i].Annotations[k].Value = s.String()
		}

		// Process labels
		for k, label := range j.Labels {
			var s strings.Builder
			if tmpl, err := tmpl.Parse(label.Value); err != nil {
				return nil, err
			} else {
				tmpl.Execute(&s, j)
			}
			result.Jobs[i].Labels[k].Value = s.String()
		}
	}

	return result, nil
}

func PrintConf(c Conf) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
