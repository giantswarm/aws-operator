package ipam

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/key"
)

type ClusterCheckerConfig struct {
	CMAClient clientset.Interface
	Logger    micrologger.Logger
}

type ClusterChecker struct {
	cmaClient clientset.Interface
	logger    micrologger.Logger
}

func NewClusterChecker(config ClusterCheckerConfig) (*ClusterChecker, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	a := &ClusterChecker{
		cmaClient: config.CMAClient,
		logger:    config.Logger,
	}

	return a, nil
}

func (c *ClusterChecker) Check(ctx context.Context, namespace string, name string) (bool, error) {
	cr, err := c.cmaClient.ClusterV1alpha1().Clusters(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return false, microerror.Mask(err)
	}

	// We check the subnet we want to ensure in the CR status. In case there is no
	// subnet tracked so far, we want to proceed with the allocation process. Thus
	// we return true.
	if key.StatusClusterNetworkCIDR(*cr) == "" {
		return true, nil
	}

	// At this point the subnet is already allocated for the CR we check here. So
	// we do not want to proceed further and return false.
	return false, nil
}
