package spec

type Calico struct {
	CIDR int `json:"cidr" yaml:"cidr"`
	// Domain is the API domain for Calico, e.g.
	// calico.<cluster-id>.g8s.fra-1.giantswarm.io.
	Domain string `json:"domain" yaml:"domain"`
	MTU    int    `json:"mtu" yaml:"mtu"`
	Subnet string `json:"subnet" yaml:"subnet"`
}
