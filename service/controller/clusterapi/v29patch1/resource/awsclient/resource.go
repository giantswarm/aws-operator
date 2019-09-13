package awsclient

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29patch1/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29patch1/credential"
)

const (
	Name = "awsclientv29patch1"
)

type Config struct {
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger
	ToClusterFunc func(v interface{}) (v1alpha1.Cluster, error)

	CPAWSConfig aws.Config
}

type Resource struct {
	k8sClient     kubernetes.Interface
	logger        micrologger.Logger
	toClusterFunc func(v interface{}) (v1alpha1.Cluster, error)

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

func (r *Resource) addAWSClientsToContext(ctx context.Context, cr v1alpha1.Cluster) error {
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
