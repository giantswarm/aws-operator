package docker

type Docker struct {
	// Image is the full qualified docker image,
	// e.g.quay.io/giantswarm/docker-kubectl:a121f8d14cd14567abc2ec20a7258be9d70ecb45
	Image string `json:"image" yaml:"image"`
}
