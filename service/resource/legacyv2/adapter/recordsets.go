package adapter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/service/keyv2"
	"github.com/giantswarm/microerror"
)

// template related to this adapter: service/templates/cloudformation/recordsets.yaml

type recordSetsAdapter struct {
	APIELBHostedZones         string
	APIELBDomain              string
	EtcdELBDNS                string
	EtcdELBHostedZones        string
	EtcdELBAliasHostedZone    string
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

	ingressELB, err := ELBDescription(clients, r.IngressELBDomain, customObject)
	if err != nil {
		return microerror.Mask(err)
	}
	r.IngressELBDNS = *ingressELB.DNSName
	r.IngressELBAliasHostedZone = *ingressELB.CanonicalHostedZoneNameID

	etcdELB, err := ELBDescription(clients, r.EtcdELBDomain, customObject)
	if err != nil {
		return microerror.Mask(err)
	}
	r.EtcdELBDNS = *etcdELB.DNSName
	r.EtcdELBAliasHostedZone = *etcdELB.CanonicalHostedZoneNameID

	return nil
}

func ELBDescription(clients Clients, domain string, customObject v1alpha1.AWSConfig) (*elb.LoadBalancerDescription, error) {
	elbName, err := keyv2.LoadBalancerName(domain, customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	input := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{
			aws.String(elbName),
		},
	}

	res, err := clients.ELB.DescribeLoadBalancers(input)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if len(res.LoadBalancerDescriptions) != 1 {
		return nil, microerror.Mask(tooManyResultsError)
	}

	return res.LoadBalancerDescriptions[0], nil
}
