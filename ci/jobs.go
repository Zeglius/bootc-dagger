package main

import (
	"dagger/ci/internal/dagger"

	"github.com/google/uuid"
)

type JobID = string

type Job struct {
	Containerfile string   `json:"containerfile"`
	BuildArgs     []string `json:"build-args,omitempty"`
	Annotations   []string `json:"annotations,omitempty"`
	OutputNames   []string `json:"output-names,omitempty"`
	OutputTags    []string `json:"output-tags,omitempty"`
}

type JobResult struct {
	ImgRef string
	Err    error
}

type JobPipeline struct {
	Id JobID
	// Contains the context associated with the container build.
	*Job
	Ctr *dagger.Container
}

func NewJobPipeline(j *Job, ctr *dagger.Container) *JobPipeline {
	return &JobPipeline{
		Job: j,
		Ctr: ctr,
		Id:  uuid.NewString(),
	}
}
