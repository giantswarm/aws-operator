package natgatewayaddresses

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v5/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	Name = "natgatewayaddresses"
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

func (r *Resource) addNATGatewayAddressesToContext(ctx context.Context, cr infrastructurev1alpha3.AWSCluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var natGatewayAddresses []*ec2.Address
	{
		i := &ec2.DescribeAddressesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String(fmt.Sprintf("tag:%s", key.TagInstallation)),
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
