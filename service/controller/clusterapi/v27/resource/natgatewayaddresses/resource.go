package natgatewayaddresses

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/controllercontext"
)

const (
	Name = "natgatewayaddressesv27"
)

type Config struct {
	Logger micrologger.Logger

	Installation string
}

type Resource struct {
	logger micrologger.Logger

	installation string
}

// New returns a resource to get all EIPs tagged with the control plane
// installation tag. Each EIP is associated with a control plane NAT gateway
// which are used to compute the tenant cluster's security group rules.
func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.Installation == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Installation must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		installation: config.Installation,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) addNATGatewayAddressesToContext(ctx context.Context, cr v1alpha1.Cluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var natGatewayAddresses []*ec2.Address
	{
		i := &ec2.DescribeAddressesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("tag:giantswarm.io/installation"),
					Values: []*string{
						aws.String(r.installation),
					},
				},
			},
		}
		o, err := cc.Client.ControlPlane.AWS.EC2.DescribeAddresses(i)
		if err != nil {
			return microerror.Mask(err)
		}

		natGatewayAddresses = o.Addresses
	}

	cc.Status.ControlPlane.NATGateway.Addresses = natGatewayAddresses

	return nil
}
