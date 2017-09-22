package clustertpr

import "github.com/giantswarm/clustertpr/spec"

type Spec struct {
	Calico     spec.Calico     `json:"calico" yaml:"calico"`
	Cluster    spec.Cluster    `json:"cluster" yaml:"cluster"`
	Customer   spec.Customer   `json:"customer" yaml:"customer"`
	Docker     spec.Docker     `json:"docker" yaml:"docker"`
	Etcd       spec.Etcd       `json:"etcd" yaml:"etcd"`
	Kubernetes spec.Kubernetes `json:"kubernetes" yaml:"kubernetes"`
	Masters    []spec.Node     `json:"masters" yaml:"masters"`
	Vault      spec.Vault      `json:"vault" yaml:"vault"`
	Version    string          `json:"version" yaml:"version"`
	Workers    []spec.Node     `json:"workers" yaml:"workers"`
}
