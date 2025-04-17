package main

type Job struct {
	// Containerfile is the path to the Dockerfile used for building the container.
	//
	// Example:
	// 	Containerfile: "Dockerfile"
	Containerfile string `json:"containerfile"`
	// BuildArgs represents a list of key-value pairs for container build arguments.
	// Each build argument consists of a Key and Value field.
	//
	// Example:
	// 	BuildArgs: []Pair{{Key: "ARG1", Value: "value1"}}
	BuildArgs []Pair `json:"build-args,omitempty"`
	// Annotations are metadata labels that can be added to the container.
	//
	// Example:
	// 	Annotations: []Pair{{Key: "KEY1", Value: "VALUE1"}}
	Annotations []Pair `json:"annotations,omitempty"`
	// Labels are metadata labels that can be added to the container.
	//
	// Example:
	// 	Labels: []Pair{{Key: "KEY1", Value: "VALUE1"}}
	Labels []Pair `json:"labels,omitempty"`
	// OutputName specifies name for the output container image.
	//
	// Example:
	// 	OutputName: "ghcr.io/ublue-os/bluefin"
	OutputName string `json:"output-name,omitempty"`
	// OutputTags specifies tags for the output container images.
	//
	// Example:
	// 	OutputTags: []string{"latest", "v1.0"}
	OutputTags []string `json:"output-tags,omitempty"`
}

type Pair struct {
	// Key is the name/identification of a build argument, annotation, or label.
	Key string
	// Value is the corresponding value associated with the Key.
	Value string
}
