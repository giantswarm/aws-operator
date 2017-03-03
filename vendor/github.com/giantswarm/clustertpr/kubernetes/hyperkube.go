package kubernetes

type Hyperkube struct {
	// Repository is the registry/namespace/repository identifier for the
	// hyperkube Docker image.
	Repository string `json:"repository" yaml:"repository"`
	// Version is the version tag for the hyperkube Docker image.
	Version string `json:"version" yaml:"version"`
}
