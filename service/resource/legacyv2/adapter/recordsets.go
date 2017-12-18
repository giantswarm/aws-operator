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
	APIELBDNS                string
	APIELBHostedZones        string
	APIELBDomain             string
	EtcdELBDNS               string
	EtcdELBHostedZones       string
	EtcdELBDomain            string
	IngressELBDNS            string
	IngressELBHostedZones    string
	IngressELBDomain         string
	IngressWildcardELBDomain string
}

func (r *recordSetsAdapter) getRecordSets(customObject v1alpha1.AWSConfig, clients Clients) error {
	r.APIELBHostedZones = customObject.Spec.AWS.API.HostedZones
	r.APIELBDomain = customObject.Spec.Cluster.Kubernetes.API.Domain
	r.EtcdELBHostedZones = customObject.Spec.AWS.Etcd.HostedZones
	r.EtcdELBDomain = customObject.Spec.Cluster.Etcd.Domain
	r.IngressELBHostedZones = customObject.Spec.AWS.Ingress.HostedZones
	r.IngressELBDomain = customObject.Spec.Cluster.Kubernetes.IngressController.Domain
	r.IngressWildcardELBDomain = customObject.Spec.Cluster.Kubernetes.IngressController.WildcardDomain

	apiDNS, err := ELBDNS(clients, r.APIELBDomain, customObject)
	if err != nil {
		return microerror.Mask(err)
	}
	r.APIELBDNS = apiDNS

	ingressDNS, err := ELBDNS(clients, r.IngressELBDomain, customObject)
	if err != nil {
		return microerror.Mask(err)
	}
	r.IngressELBDNS = ingressDNS

	etcdDNS, err := ELBDNS(clients, r.EtcdELBDomain, customObject)
	if err != nil {
		return microerror.Mask(err)
	}
	r.EtcdELBDNS = etcdDNS

	return nil
}

func ELBDNS(clients Clients, domain string, customObject v1alpha1.AWSConfig) (string, error) {
	elbName, err := keyv2.LoadBalancerName(domain, customObject)
	if err != nil {
		return "", microerror.Mask(err)
	}

	input := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{
			aws.String(elbName),
		},
	}

	res, err := clients.ELB.DescribeLoadBalancers(input)
	if err != nil {
		return "", microerror.Mask(err)
	}
	if len(res.LoadBalancerDescriptions) != 1 {
		return "", microerror.Mask(tooManyResultsError)
	}

	return *res.LoadBalancerDescriptions[0].DNSName, nil
}
