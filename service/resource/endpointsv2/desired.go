package endpointsv2

import (
	"context"

	"github.com/giantswarm/microerror"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/api/core/v1"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv2.ToCustomObject(obj)
	if IsNotFound(err) {
		// Fall through.
		return nil, nil
	}
	if err != nil {
		return nil, microerror.Mask(err)
	}

	instanceName := keyv2.MasterInstanceName(customObject)
	masterInstance, err := r.findMasterInstance(instanceName)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	endpoints := &v1.Endpoints{
		ObjectMeta: apismetav1.ObjectMeta{
			Name:      masterEndpointsName,
			Namespace: keyv2.ClusterID(customObject),
			Labels: map[string]string{
				"app":      masterEndpointsName,
				"cluster":  keyv2.ClusterID(customObject),
				"customer": keyv2.CustomerID(customObject),
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
