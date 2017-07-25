package hyperkube

type Docker struct {
	// Image is the full qualified docker image,
	// e.g.quay.io/coreos/hyperkube:v1.5.2_coreos.2.
	Image string `json:"image" yaml:"image"`
}
