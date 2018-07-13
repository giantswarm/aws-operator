package ipam

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/service/controller/v14/key"
	"github.com/giantswarm/microerror"
)

// EnsureCreated allocates guest cluster network segment. It gathers existing
// subnets from existing AWSConfig/Status objects and existing VPCs from AWS.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "checking if subnet needs to be allocated for cluster")

	if key.ClusterNetworkCIDR(customObject) == "" {

		r.logger.LogCtx(ctx, "level", "debug", "message", "allocating subnet for cluster")

		subnetCIDR, err := allocateSubnet(customObject)
		if err != nil {
			return microerror.Mask(err)
		}

		customObject.Status.Cluster.Network.CIDR = subnetCIDR
		_, err = r.g8sClient.ProviderV1alpha1().AWSConfigs(customObject.Namespace).UpdateStatus(&customObject)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("subnet %s allocated for cluster", subnetCIDR))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "subnet doesn't need to be allocated for cluster")
	}

	return nil
}

func allocateSubnet(customObject v1alpha1.AWSConfig) (string, error) {
	// TODO: Implement
	return "", nil
}
