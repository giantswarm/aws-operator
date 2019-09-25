package peerrolearn

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/legacy/v31/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v31/key"
)

const (
	Name = "peerrolearnv31"
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

func (r *Resource) addPeerRoleARNToContext(ctx context.Context, cr v1alpha1.AWSConfig) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var peerRoleArn string
	{
		i := &iam.GetRoleInput{
			RoleName: aws.String(key.PeerAccessRoleName(cr)),
		}
		o, err := cc.Client.ControlPlane.AWS.IAM.GetRole(i)
		if IsNotFound(err) {
			return microerror.Maskf(notFoundError, key.PeerAccessRoleName(cr))
		} else if err != nil {
			return microerror.Mask(err)
		}

		peerRoleArn = *o.Role.Arn
	}

	cc.Status.ControlPlane.PeerRole.ARN = peerRoleArn

	return nil
}
