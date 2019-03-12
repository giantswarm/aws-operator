package vpccidr

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/v25/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v25/key"
	"github.com/giantswarm/aws-operator/service/vpccidr"
)

const (
	Name = "vpccidrv25"
)

type Config struct {
	Logger micrologger.Logger
}

type Resource struct {
	logger micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) addVPCCIDRToContext(ctx context.Context, cr v1alpha1.AWSConfig) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var vpcCIDRService *vpccidr.VPCCIDR
	{
		c := vpccidr.Config{
			EC2:    cc.Client.ControlPlane.AWS.EC2,
			Logger: r.logger,
		}

		vpcCIDRService, err = vpccidr.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	cidr, err := vpcCIDRService.Lookup(ctx, key.PeerID(cr))
	if err != nil {
		return microerror.Mask(err)
	}

	cc.Status.ControlPlane.VPC.CIDR = cidr

	return nil
}
