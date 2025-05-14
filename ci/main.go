// Build and distribute Bootc container images across registries in a standarized way.
package main

import (
	"context"
	"dagger/ci/config"
	"dagger/ci/internal/dagger"
	"dagger/ci/tmpls"
	"dagger/ci/types/syncmap"
	"dagger/ci/utils/confparser"
	"fmt"
	"strings"
	"text/template"

	"github.com/invopop/jsonschema"
	"github.com/stoewer/go-strcase"
	"golang.org/x/sync/errgroup"
)

type Ci struct{}

type Builder struct {
	Conf         config.ConfString
	BuildContext *dagger.Directory
}

// Start the CI pipeline
func (m *Ci) NewBuilder(
	ctx context.Context,
	// +defaultPath="./bootc-ci.yaml"
	cfgFile *dagger.File,
	// +defaultPath="."
	buildContext *dagger.Directory,
	// +optional
	dryRun bool, // Skip publishing
	// +optional
	secrets []*dagger.Secret,
) (*Builder, error) {

	builder := &Builder{
		Conf:         "",
		BuildContext: buildContext,
	}

	if cs, err := m.parseConfFile(ctx, cfgFile, secrets...); err != nil {
		return nil, err
	} else {
		if cs == "" {
			return nil, fmt.Errorf("Configuration file didnt load correctly")
		}

		c, err := config.ReadConfString(cs)
		if err != nil {
			return nil, err
		}

		if len(c.Jobs) == 0 {
			return nil, fmt.Errorf("There are no jobs in the config: %s", c)
		}

		builder.Conf, err = c.ToConfString()
		if err != nil {
			return nil, err
		}
	}

	return builder, nil
}

func (b *Builder) Build(
	ctx context.Context,
	// +optional
	dryRun bool,
) ([]string, error) {
	ctrs := syncmap.New[int, []string]()
	eg, gctx := errgroup.WithContext(ctx)
	var conf config.Conf
	if c, err := config.ReadConfString(b.Conf); err != nil {
		return nil, err
	} else {
		conf = c
	}

	for i, j := range conf.Jobs {
		eg.Go(func() error {
			var (
				ctr *dagger.Container = nil
				err error             = nil
			)
			ctr = buildContainer(j, b.BuildContext)
			ctr = labelAndAnnotate(j, ctr)

			refs := make([]string, len(conf.Jobs))
			if !dryRun {
				refs, err = publishImages(gctx, j, ctr)
				if err != nil {
					return err
				}
			}

			ctrs.Store(i, refs)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	imgRefs := []string{}
	ctrs.Range(func(key int, value []string) bool {
		imgRefs = append(imgRefs, value...)
		return true
	})
	return imgRefs, nil
}

// Build a container image using the provided job configuration like build-args.
func buildContainer(j config.Job, d *dagger.Directory) *dagger.Container {
	buildOpts := dagger.ContainerBuildOpts{}
	if j.Containerfile != "" {
		buildOpts.Dockerfile = j.Containerfile
	} else {
		buildOpts.Dockerfile = "Dockerfile"
	}

	// Set build arguments
	if len(j.BuildArgs) != 0 {
		for k, v := range j.BuildArgs {
			// build-args is a name=value string, so we need to split
			buildOpts.BuildArgs = append(
				buildOpts.BuildArgs,
				dagger.BuildArg{Name: k, Value: v},
			)
		}
	}

	// Build image
	ctr := dag.Container().Build(d, buildOpts)
	return ctr
}

// Publish container images with the provided tags to a remote image
// registry. The provided container is published with each output tag using the output
// image name and returns the references to each published image.
func publishImages(ctx context.Context, j config.Job, ctr *dagger.Container) ([]string, error) {
	var imgRefs []string
	for _, t := range j.OutputTags {
		im, err := ctr.
			Publish(
				ctx,
				strings.ToLower(j.OutputName+":"+t),
			)
		if err != nil {
			return nil, err
		}
		imgRefs = append(imgRefs, im)
	}

	return imgRefs, nil
}

// labelAndAnnotate labels and annotations to container images.
// Annotations and labels take the form of "name=value" strings.
func labelAndAnnotate(j config.Job, ctr *dagger.Container) *dagger.Container {
	// Add annotations
	for k, v := range j.Annotations {
		ctr = ctr.WithAnnotation(k, v)
	}

	// Add labels
	for k, v := range j.Labels {
		ctr = ctr.WithLabel(k, v)
	}
	return ctr
}

func (m *Ci) parseConfFile(ctx context.Context, cfgFile *dagger.File, secrets ...*dagger.Secret) (config.ConfString, error) {

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

	confStr, err := confparser.Parse(cfgContents, confparser.ParseOpts{
		TmplName: cfgFileName,
		TmplFuncs: []template.FuncMap{
			tmpls.TmplFuncs,
			tmpls.TmpFuncsWithCtr(ctx, dag),
		},
		Data: map[string]any{
			"secrets": tmpls.SecretsToMap(ctx, secrets),
		},
	})

	if err != nil {
		return "", err
	}

	return confStr, nil
}

// Returns the json schema that the config file follows.
func (m *Ci) ConfigJsonSchema() string {
	r := &jsonschema.Reflector{
		KeyNamer:       strcase.KebabCase,
		ExpandedStruct: true,
	}

	json, _ := r.Reflect((*config.Conf)(nil)).MarshalJSON()
	return string(json)
}
