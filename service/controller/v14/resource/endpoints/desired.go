package endpoints

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/v13/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	instanceName := key.MasterInstanceName(customObject)
	masterInstance, err := r.findMasterInstance(ctx, instanceName)
	if IsNotFound(err) {
		// During updates the master instance is shut down and thus cannot be found.
		// In such cases we cancel the reconciliation for the endpoint resource.
		// This should be ok since all endpoints should be created and up to date
		// already. In case we miss an update it will be done on the next resync
		// period once the master instance is up again.
		//
		// TODO we might want to alert at some point when the master instance was
		// not seen for too long. Like we should be able to find it again after
		// three resync periods max or something.
		r.logger.LogCtx(ctx, "level", "debug", "message", "master instance not found")
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")

		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	endpoints := &v1.Endpoints{
		ObjectMeta: apismetav1.ObjectMeta{
			Name:      masterEndpointsName,
			Namespace: key.ClusterID(customObject),
			Labels: map[string]string{
				"app":      masterEndpointsName,
				"cluster":  key.ClusterID(customObject),
				"customer": key.CustomerID(customObject),
			},
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: *masterInstance.PrivateIpAddress,
					},
				},
				Ports: []v1.EndpointPort{
					{
						Port: httpsPort,
					},
				},
			},
		},
	}

	return endpoints, nil
}
