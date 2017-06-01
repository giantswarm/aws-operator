package docker

type Docker struct {
	// Image is the full qualified docker image.
	// e.g. rlister/amazon-ssm-agent
	Image string `json:"image" yaml:"image"`
}
