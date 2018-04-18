package endpoints

import (
	"context"

	"github.com/giantswarm/microerror"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/v5/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if IsNotFound(err) {
		// Fall through.
		return nil, nil
	}
	if err != nil {
		return nil, microerror.Mask(err)
	}

	instanceName := key.MasterInstanceName(customObject)
	masterInstance, err := r.findMasterInstance(instanceName)
	if err != nil {
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
