// Build and distribute Bootc container images across registries in a standarized way.
package main

import (
	"context"
	"dagger/ci/internal/dagger"
	"dagger/ci/types/syncmap"
	"encoding/json"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type Ci struct{}

type Builder struct {
	Conf         ConfString
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
) (*Builder, error) {

	builder := &Builder{
		Conf:         "",
		BuildContext: buildContext,
	}

	if c, err := m.parseConfFile(ctx, cfgFile); err != nil {
		return nil, err
	} else {
		if c == "" {
			return nil, fmt.Errorf("Configuration file didnt load correctly")
		}

		{
			var s Conf
			if err := json.Unmarshal([]byte(c), &s); err != nil {
				return nil, err
			}
			if len(s.Jobs) == 0 {
				return nil, fmt.Errorf("There are no jobs in the config: %s", c)
			}
		}

		builder.Conf = c
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
	var conf Conf
	if err := json.Unmarshal([]byte(b.Conf), &conf); err != nil {
		return nil, err
	}

	for i, j := range conf.Jobs {
		eg.Go(func() error {
			var (
				ctr *dagger.Container = nil
				err error             = nil
			)
			ctr = buildContainer(j, b.BuildContext)
			ctr = labelAndAnnotate(j, ctr)
			// Necessary in order to trigger the container BuildContext
			// inside the coroutine.
			ctr, err = ctr.Sync(gctx)
			if err != nil {
				return err
			}

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
func buildContainer(j Job, d *dagger.Directory) *dagger.Container {
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
func publishImages(ctx context.Context, j Job, ctr *dagger.Container) ([]string, error) {
	var imgRefs []string
	for _, t := range j.OutputTags {
		im, err := ctr.
			Publish(ctx, j.OutputName+":"+t)
		if err != nil {
			return nil, err
		}
		imgRefs = append(imgRefs, im)
	}

	return imgRefs, nil
}

// labelAndAnnotate labels and annotations to container images.
// Annotations and labels take the form of "name=value" strings.
func labelAndAnnotate(j Job, ctr *dagger.Container) *dagger.Container {
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
