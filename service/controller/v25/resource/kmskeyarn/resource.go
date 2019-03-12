package kmskeyarn

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/v24/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v24/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v24/key"
)

const (
	Name = "kmskeyarnv24"
)

type Config struct {
	Logger micrologger.Logger

	EncrypterBackend string
}

type Resource struct {
	logger micrologger.Logger

	encrypterBackend string
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.EncrypterBackend == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.EncrypterBackend must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		encrypterBackend: config.EncrypterBackend,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) addKMSKeyARNToContext(ctx context.Context, cr v1alpha1.AWSConfig) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if r.encrypterBackend == encrypter.KMSBackend {
		a := fmt.Sprintf("alias/%s", key.ClusterID(cr))

		i := &kms.DescribeKeyInput{
			KeyId: aws.String(a),
		}

		o, err := cc.Client.TenantCluster.AWS.KMS.DescribeKey(i)
		if IsNotFound(err) {
			return microerror.Maskf(notFoundError, a)
		} else if err != nil {
			return microerror.Mask(err)
		}

		cc.Status.TenantCluster.KMS.KeyARN = *o.KeyMetadata.Arn
	}

	return nil
}
