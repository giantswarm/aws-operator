package legacy

import (
	"fmt"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/awstpr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/microerror"
)

type MasterServiceInput struct {
	Clients  awsutil.Clients
	Cluster  awstpr.CustomObject
	MasterID string
}

func (s *Resource) createMasterService(input MasterServiceInput) error {
	instances, err := aws.FindInstances(aws.FindInstancesInput{
		Clients: input.Clients,
		Logger:  s.logger,
		Pattern: instanceName(instanceNameInput{
			clusterName: key.ClusterID(input.Cluster),
			prefix:      prefixMaster,
			no:          0,
		}),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	if len(instances) < 1 {
		return microerror.Maskf(notFoundError, "could not find instance: %s", input.MasterID)
	}
	if len(instances) > 1 {
		return microerror.Maskf(tooManyResultsError, "too many instances: %s", input.MasterID)
	}

	masterInstance := instances[0]

	service := v1.Service{
		ObjectMeta: apismetav1.ObjectMeta{
			Name:      "master",
			Namespace: key.ClusterID(input.Cluster),
			Labels: map[string]string{
				"app":      "master",
				"cluster":  key.ClusterID(input.Cluster),
				"customer": key.CustomerID(input.Cluster),
			},
			Annotations: map[string]string{
				"giantswarm.io/prometheus-cluster": key.ClusterID(input.Cluster),
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Protocol:   v1.ProtocolTCP,
					Port:       443,
					TargetPort: intstr.FromInt(443),
				},
			},
		},
	}

	endpoint := v1.Endpoints{
		ObjectMeta: apismetav1.ObjectMeta{
			Name:      "master",
			Namespace: key.ClusterID(input.Cluster),
			Labels: map[string]string{
				"app":      "master",
				"cluster":  key.ClusterID(input.Cluster),
				"customer": key.CustomerID(input.Cluster),
			},
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: masterInstance.PrivateIpAddress,
					},
				},
				Ports: []v1.EndpointPort{
					{
						Port: 443,
					},
				},
			},
		},
	}

	if _, err := s.k8sClient.Core().Services(service.ObjectMeta.Namespace).Create(&service); err != nil && !apierrors.IsAlreadyExists(err) {
		return microerror.Mask(err)
	}

	if _, err := s.k8sClient.Core().Endpoints(endpoint.ObjectMeta.Namespace).Create(&endpoint); err != nil && !apierrors.IsAlreadyExists(err) {
		return microerror.Mask(err)
	}

	s.logger.Log("info", fmt.Sprintf("created master service for scraping"))

	return nil
}
