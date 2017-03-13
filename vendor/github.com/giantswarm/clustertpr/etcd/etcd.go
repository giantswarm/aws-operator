package etcd

type Etcd struct {
	Domain string `json:"domain" yaml:"domain"`
	Port   int    `json:"port" yaml:"port"`
	Prefix string `json:"prefix" yaml:"prefix"`
}
