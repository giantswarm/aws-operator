package awstpr

import (
	"github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/awstpr/spec/calico"
	"github.com/giantswarm/awstpr/spec/cloudflare"
	"github.com/giantswarm/awstpr/spec/cluster"
	"github.com/giantswarm/awstpr/spec/customer"
	"github.com/giantswarm/awstpr/spec/docker"
	"github.com/giantswarm/awstpr/spec/etcd"
	"github.com/giantswarm/awstpr/spec/kubernetes"
	"github.com/giantswarm/awstpr/spec/node"
	"github.com/giantswarm/awstpr/spec/operator"
	"github.com/giantswarm/awstpr/spec/vault"
)

type Spec struct {
	Aws        aws.Aws               `json:"aws"`
	Calico     calico.Calico         `json:"calico"`
	Cloudflare cloudflare.Cloudflare `json:"cloudflare"`
	Cluster    cluster.Cluster       `json:"cluster"`
	Customer   customer.Customer     `json:"customer"`
	Docker     docker.Docker         `json:"docker"`
	Etcd       etcd.Etcd             `json:"etcd"`
	Kubernetes kubernetes.Kubernetes `json:"kubernetes"`
	Masters    []node.Node           `json:"masters"`
	Operator   operator.Operator     `json:"operator"`
	Vault      vault.Vault           `json:"vault"`
	Workers    []node.Node           `json:"workers"`
}
