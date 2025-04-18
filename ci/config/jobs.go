package config

type Job struct {
	// Containerfile is the path to the Dockerfile used for building the container.
	//
	// Example:
	// 	Containerfile: "Dockerfile"
	Containerfile string `json:"containerfile"`
	// BuildArgs represents a map of key-value pairs for container build arguments.
	// Each build argument consists of a Key and Value field.
	//
	// Example:
	// 	BuildArgs: map[string]string{"ARG1": "value1"}
	BuildArgs map[string]string `json:"build-args,omitempty"`
	// Annotations are metadata labels that can be added to the container.
	//
	// Example:
	// 	Annotations: map[string]string{"KEY1": "VALUE1"}
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels are metadata labels that can be added to the container.
	//
	// Example:
	// 	Labels: map[string]string{"KEY1": "VALUE1"}
	Labels map[string]string `json:"labels,omitempty"`
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
