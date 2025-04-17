// Build and distribute Bootc container images across registries in a standarized way.
package main

import (
	"context"
	"dagger/ci/internal/dagger"
	"dagger/ci/types/syncmap"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type Ci struct {
	Conf         *Conf
	BuildContext *dagger.Directory // Contains the context of our CI pipeline execution
}

func New(
	ctx context.Context,
	// +defaultPath="./bootc-ci.yaml"
	cfgFile *dagger.File,
	// +defaultPath="."
	buildContext *dagger.Directory,
) (*Ci, error) {
	res := &Ci{BuildContext: buildContext}

	conf, err := res.parseConfFile(ctx, cfgFile)
	if err != nil {
		return nil, err
	}
	res.Conf = conf

	return res, nil
}

// Start the CI pipeline
func (m *Ci) Run(
	// +optional
	dryRun bool, // Skip publishing
) ([]string, error) {
	if len(m.Conf.Jobs) == 0 {
		return nil, fmt.Errorf("There are no jobs in the config: %v", *m.Conf)
	}

	ctrs := syncmap.New[int, []string]()
	eg, gctx := errgroup.WithContext(context.Background())
	for i, j := range m.Conf.Jobs {
		eg.Go(func() error {
			var (
				ctr *dagger.Container = nil
				err error             = nil
			)
			ctr = m.buildContainer(j)
			ctr, err = labelAndAnnotate(j, ctr).
				// Necessary in order to trigger the container BuildContext
				// inside the coroutine.
				Sync(gctx)
			if err != nil {
				return err
			}

			refs := make([]string, len(m.Conf.Jobs))
			if !dryRun {
				refs, err = m.PublishImages(gctx, j, ctr)
				if err != nil {
					return err
				}
			}

			ctrs.Store(i, refs)
			return nil
		})
	}

	err := eg.Wait()
	if err != nil {
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
func (m *Ci) buildContainer(j Job) *dagger.Container {
	buildOpts := dagger.ContainerBuildOpts{}
	if j.Containerfile != "" {
		buildOpts.Dockerfile = j.Containerfile
	} else {
		buildOpts.Dockerfile = "Dockerfile"
	}

	// Set build arguments
	if len(j.BuildArgs) != 0 {
		for _, ba := range j.BuildArgs {
			// build-args is a name=value string, so we need to split
			buildOpts.BuildArgs = append(
				buildOpts.BuildArgs,
				dagger.BuildArg{Name: ba.Key, Value: ba.Value},
			)
		}
	}

	// Build image
	ctr := dag.Container().Build(m.BuildContext, buildOpts)
	return ctr
}

// Publish container images with the provided tags to a remote image
// registry. The provided container is published with each output tag using the output
// image name and returns the references to each published image.
func (m *Ci) PublishImages(ctx context.Context, j Job, ctr *dagger.Container) ([]string, error) {
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
	for _, a := range j.Annotations {
		ctr = ctr.WithAnnotation(a.Key, a.Value)
	}

	// Add labels
	for _, l := range j.Labels {
		ctr = ctr.WithLabel(l.Key, l.Value)
	}
	return ctr
}
