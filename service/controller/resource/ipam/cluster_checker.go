package ipam

import (
	"context"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v5/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/key"

	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterCheckerConfig struct {
	CtrlClient ctrlClient.Client
	Logger     micrologger.Logger
}

type ClusterChecker struct {
	ctrlClient ctrlClient.Client
	logger     micrologger.Logger
}

func NewClusterChecker(config ClusterCheckerConfig) (*ClusterChecker, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	a := &ClusterChecker{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return a, nil
}

func (c *ClusterChecker) Check(ctx context.Context, namespace string, name string) (bool, error) {
	var cr *infrastructurev1alpha3.AWSCluster
	err := c.ctrlClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, cr)
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
