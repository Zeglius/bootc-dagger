// Build and distribute Bootc container images across registries
// in a sane way.
package main

import (
	"context"
	"dagger/ci/internal/dagger"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
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
func (m *Ci) Run(ctx context.Context) (imgRefs []string, err error) {
	cfg := m.Conf
	if len(cfg.Jobs) == 0 {
		return []string{}, fmt.Errorf("There are no jobs in the config")
	}

	// Start a corroutine per job
	wg := new(sync.WaitGroup)
	jobsResChan := make(chan JobResult, len(cfg.Jobs))
	jobs := []*JobPipeline{}
	for _, j := range cfg.Jobs {
		jobs = append(jobs, NewJobPipeline(&j, dag.Container()))
		wg.Add(1)
	}

	for _, j := range jobs {
		go func(c chan<- JobResult, wg *sync.WaitGroup) {
			c <- m.runJob(ctx, j)
			wg.Done()
		}(jobsResChan, wg)
	}

	select {
	case v := <-jobsResChan:
		if err := v.Err; err != nil {
			return []string{}, err
		} else {
			imgRefs = append(imgRefs, v.ImgRef)
		}
	}

	wg.Wait()

	return imgRefs, nil
}

// TODO(Zeg): Parse annotations fields as Go templates
func (m *Ci) runJob(ctx context.Context, j *JobPipeline) JobResult {
	// Prepare build options
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

	// Add annotations
	for _, a := range j.Annotations {
		k, v, _ := strings.Cut(a, "=")
		ctr = ctr.WithAnnotation(k, v)
	}

	imgRef, err := ctr.
		Publish(ctx, fmt.Sprintf("ttl.sh/%s:latest", uuid.NewString()))
	return JobResult{ImgRef: imgRef, Err: err}
}
