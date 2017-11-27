package cloudformation

import (
	"fmt"
	"strings"

	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"
)

// LoadBalancerName produces a unique name for the load balancer.
// It takes the domain name, extracts the first subdomain, and combines it with the cluster name.
func LoadBalancerName(domainName string, cluster awstpr.CustomObject) (string, error) {
	if key.ClusterID(cluster) == "" {
		return "", microerror.Maskf(missingCloudConfigKeyError, "spec.cluster.cluster.id")
	}

	componentName, err := componentName(domainName)
	if err != nil {
		return "", microerror.Maskf(malformedCloudConfigKeyError, "spec.cluster.cluster.id")
	}

	lbName := fmt.Sprintf("%s-%s", key.ClusterID(cluster), componentName)

	return lbName, nil
}

// componentName returns the first component of a domain name.
// e.g. apiserver.example.customer.cloud.com -> apiserver
func componentName(domainName string) (string, error) {
	splits := strings.SplitN(domainName, ".", 2)

	if len(splits) != 2 {
		return "", microerror.Mask(malformedCloudConfigKeyError)
	}

	return splits[0], nil
}

// ingressLoadBalancerName returns the name of the ingress load balancer
func ingressLoadBalancerName(customObject awstpr.CustomObject) (string, error) {
	return LoadBalancerName(key.IngressDomain(customObject), customObject)
}
