package adapter

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

// template related to this adapter: service/templates/cloudformation/recordsets.yaml

type recordSetsAdapter struct {
	APIELBHostedZones         string
	APIELBDomain              string
	EtcdELBHostedZones        string
	EtcdELBDomain             string
	IngressELBDNS             string
	IngressELBHostedZones     string
	IngressELBAliasHostedZone string
	IngressELBDomain          string
	IngressWildcardELBDomain  string
}

func (r *recordSetsAdapter) getRecordSets(customObject v1alpha1.AWSConfig, clients Clients) error {
	r.APIELBHostedZones = customObject.Spec.AWS.API.HostedZones
	r.APIELBDomain = customObject.Spec.Cluster.Kubernetes.API.Domain
	r.EtcdELBHostedZones = customObject.Spec.AWS.Etcd.HostedZones
	r.EtcdELBDomain = customObject.Spec.Cluster.Etcd.Domain
	r.IngressELBHostedZones = customObject.Spec.AWS.Ingress.HostedZones
	r.IngressELBDomain = customObject.Spec.Cluster.Kubernetes.IngressController.Domain
	r.IngressWildcardELBDomain = customObject.Spec.Cluster.Kubernetes.IngressController.WildcardDomain

	return nil
}
