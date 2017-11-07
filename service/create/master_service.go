package create

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/awstpr"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/microerror"
)

type MasterServiceInput struct {
	Clients   awsutil.Clients
	Cluster   awstpr.CustomObject
	MasterIDs []string
}

func (s *Service) createMasterService(input MasterServiceInput) error {
	for _, masterID := range input.MasterIDs {
		reservations, err := input.Clients.EC2.DescribeInstances(&ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("instance-id"),
					Values: []*string{
						aws.String(masterID),
					},
				},
			},
		})
		if err != nil {
			return microerror.Mask(err)
		}

		if len(reservations.Reservations) < 1 {
			return microerror.Maskf(notFoundError, "could not find master reservation: %s", masterID)
		}
		if len(reservations.Reservations) > 1 {
			return microerror.Maskf(tooManyResultsError, "too many master reservations: %s", masterID)
		}

		if len(reservations.Reservations[0].Instances) < 1 {
			return microerror.Maskf(notFoundError, "could not find master instance: %s", masterID)
		}
		if len(reservations.Reservations[0].Instances) > 1 {
			return microerror.Maskf(tooManyResultsError, "too many master instances: %s", masterID)
		}

		masterInstance := reservations.Reservations[0].Instances[0]

		service := v1.Service{
			ObjectMeta: apismetav1.ObjectMeta{
				Name:      "master",
				Namespace: key.ClusterID(input.Cluster),
				Labels: map[string]string{
					"app":      "master",
					"cluster":  key.ClusterID(input.Cluster),
					"customer": key.ClusterID(input.Cluster),
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
					"customer": key.ClusterID(input.Cluster),
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
							Port: 443,
						},
					},
				},
			},
		}

		if _, err := s.k8sClient.Core().Services(service.ObjectMeta.Namespace).Create(&service); err != nil {
			return microerror.Mask(err)
		}

		if _, err := s.k8sClient.Core().Endpoints(endpoint.ObjectMeta.Namespace).Create(&endpoint); err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
