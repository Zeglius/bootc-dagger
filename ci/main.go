// Build and distribute Bootc container images across registries
// in a sane way.
package main

import (
	"context"
	"dagger/ci/internal/dagger"
	"fmt"
	"strings"
	"sync"
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
func (m *Ci) Run(ctx context.Context) ([]string, error) {
	if len(m.Conf.Jobs) == 0 {
		return nil, fmt.Errorf("There are no jobs in the config: %v", *m.Conf)
	}

	var wg sync.WaitGroup
	results := make(chan struct {
		refs []string
		err  error
	}, len(m.Conf.Jobs))

	for _, j := range m.Conf.Jobs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			refs, err := m.runJob(ctx, j)
			results <- struct {
				refs []string
				err  error
			}{refs, err}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	imgRefs := []string{}
	for result := range results {
		if result.err != nil {
			return nil, result.err
		}
		imgRefs = append(imgRefs, result.refs...)
	}

	return imgRefs, nil
}

// Concurrent worker that builds, annotates and publish an image
// per job.
func (m *Ci) runJob(ctx context.Context, j Job) ([]string, error) {
	// Prepare build options
	ctr := m.buildContainer(j)

	// Add annotations and labels
	ctr = labelAndAnnotate(j, ctr)

	// Publish the image, per each tag
	return publishImages(ctx, j, ctr)
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
			k, v, _ := strings.Cut(ba, "=")
			buildOpts.BuildArgs = append(
				buildOpts.BuildArgs,
				dagger.BuildArg{Name: k, Value: v},
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
	for _, a := range j.Annotations {
		k, v, _ := strings.Cut(a, "=")
		ctr = ctr.WithAnnotation(k, v)
	}

	// Add labels
	for _, l := range j.Labels {
		k, v, _ := strings.Cut(l, "=")
		ctr = ctr.WithLabel(k, v)
	}
	return ctr
}
