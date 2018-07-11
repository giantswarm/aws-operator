package cloudformation

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v14/key"
)

type validator func(v1alpha1.AWSConfig) error

func (r *Resource) validateCluster(cluster v1alpha1.AWSConfig) error {
	validators := []validator{
		r.validateHostPeeringRoutes,
	}

	for _, v := range validators {
		if err := v(cluster); err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Resource) validateHostPeeringRoutes(cluster v1alpha1.AWSConfig) error {
	input := &ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("route.destination-cidr-block"),
				Values: []*string{
					aws.String(key.PrivateSubnetCIDR(cluster)),
				},
			},
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(key.PeerID(cluster)),
				},
			},
		},
	}
	output, err := r.hostClients.EC2.DescribeRouteTables(input)
	if err == nil && len(output.RouteTables) > 0 {
		return microerror.Maskf(alreadyExistsError, "route: %s", key.PrivateSubnetCIDR(cluster))
	}

	return nil
}
