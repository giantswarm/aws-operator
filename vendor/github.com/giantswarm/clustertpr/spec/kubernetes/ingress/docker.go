package ingress

type Docker struct {
	// Image is the full qualified docker image,
	// e.g. quay.io/giantswarm/nginx-ingress-controller/0.9.0-beta.11
	Image string `json:"image" yaml:"image"`
}
