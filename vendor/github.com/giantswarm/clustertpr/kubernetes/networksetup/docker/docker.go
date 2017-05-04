package docker

type Docker struct {
	// Image is the full qualified docker image,
	// e.g. giantswarm/k8s-setup-network-environment:ba2b57155d859a1fc5d378c2a09a77d7c2c755ed
	Image string `json:"image" yaml:"image"`
}
