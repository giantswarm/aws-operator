package awsclient

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/internal/credential"
)

const (
	Name = "awsclient"
)

type Config struct {
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger
	ToClusterFunc func(v interface{}) (infrastructurev1alpha2.AWSCluster, error)

	CPAWSConfig aws.Config
}

type Resource struct {
	k8sClient     kubernetes.Interface
	logger        micrologger.Logger
	toClusterFunc func(v interface{}) (infrastructurev1alpha2.AWSCluster, error)

	cpAWSConfig aws.Config
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	r := &Resource{
		k8sClient:     config.K8sClient,
		logger:        config.Logger,
		toClusterFunc: config.ToClusterFunc,

		cpAWSConfig: config.CPAWSConfig,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) addAWSClientsToContext(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		c := r.cpAWSConfig

		clients, err := aws.NewClients(c)
		if err != nil {
			return microerror.Mask(err)
		}

		cc.Client.ControlPlane.AWS = clients
	}

	{
		arn, err := credential.GetARN(r.k8sClient, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		c := r.cpAWSConfig
		c.RoleARN = arn

		clients, err := aws.NewClients(c)
		if err != nil {
			return microerror.Mask(err)
		}

		cc.Client.TenantCluster.AWS = clients
	}

	return nil
}
