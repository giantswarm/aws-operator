package spec

type Etcd struct {
	// AltNames is the alternative names used to generate certificates for Etcd.
	AltNames string `json:"altNames" yaml:"altNames"`
	// Domain is the API domain for etcd, e.g.
	// etcd.<cluster-id>.g8s.fra-1.giantswarm.io.
	Domain string `json:"domain" yaml:"domain"`
	Port   int    `json:"port" yaml:"port"`
	Prefix string `json:"prefix" yaml:"prefix"`
}
