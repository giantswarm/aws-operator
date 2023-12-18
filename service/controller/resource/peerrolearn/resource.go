package peerrolearn

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/v15/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v15/service/controller/key"
)

const (
	Name = "peerrolearn"
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

func (r *Resource) addPeerRoleARNToContext(ctx context.Context, cr infrastructurev1alpha3.AWSCluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var peerRoleArn string
	{
		i := &iam.GetRoleInput{
			RoleName: aws.String(key.RolePeerAccess(cr)),
		}
		o, err := cc.Client.ControlPlane.AWS.IAM.GetRole(i)
		if IsNotFound(err) {
			return microerror.Maskf(notFoundError, key.RolePeerAccess(cr))
		} else if err != nil {
			return microerror.Mask(err)
		}

		peerRoleArn = *o.Role.Arn
	}

	cc.Status.ControlPlane.PeerRole.ARN = peerRoleArn

	return nil
}
