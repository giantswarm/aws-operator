package ensurecpcrs

import (
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "ensurecpcrs"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

// New provides an operatorkit resource implementation in order to ensure
// existing Node Pool clusters have proper G8sControlPlane CRs on AWS. This
// includes AWSControlPlane CRs which are linked in the infrastructure reference
// of the G8sControlPlane CR. Note that this resource implementation does only
// exist for migration purposes and can be removed in major versions 9.x.x of
// aws-operator.
//
//     https://github.com/giantswarm/giantswarm/issues/9172
//
func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
