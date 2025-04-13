package main

import (
	"dagger/ci/internal/dagger"
)

type JobID = string

type Job struct {
	// Containerfile is the path to the Dockerfile used for building the container.
	//
	// Example:
	// 	Containerfile: "Dockerfile"
	Containerfile string `json:"containerfile"`
	// BuildArgs are the build arguments passed to the container build process.
	//
	// Example:
	// 	BuildArgs: []string{"ARG1=value1", "ARG2=value2"}
	BuildArgs []string `json:"build-args,omitempty"`
	// Annotations are metadata labels that can be added to the container.
	//
	// Example:
	// 	Annotations: []string{"key1=value1", "key2=value2"}
	Annotations []string `json:"annotations,omitempty"`
	// OutputNames specifies names for the output container images.
	//
	// Example:
	// 	OutputNames: []string{"ghcr.io/ublue-os/bluefin"}
	OutputNames []string `json:"output-names,omitempty"`
	// OutputTags specifies tags for the output container images.
	//
	// Example:
	// 	OutputTags: []string{"latest", "v1.0"}
	OutputTags []string `json:"output-tags,omitempty"`
}

type JobPipeline struct {
	// Contains the context associated with the container build.
	*Job
	Ctr *dagger.Container
}
