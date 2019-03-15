package tenantcluster

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Config represents the configuration used to create a new tenant cluster
// service.
type Config struct {
	CertsSearcher certs.Interface
	Logger        micrologger.Logger

	CertID          certs.Cert
	TillerNamespace string
}

// TenantCluster provides functionality for connecting to tenant clusters.
type TenantCluster struct {
	certsSearcher certs.Interface
	logger        micrologger.Logger

	certID          certs.Cert
	tillerNamespace string
}

// New creates a new tenant cluster service.
func New(config Config) (*TenantCluster, error) {
	if config.CertsSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CertsSearcher must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.CertID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.CertID must not be empty", config)
	}
	if config.TillerNamespace == "" {
		config.TillerNamespace = TillerDefaultNamespace
	}

	g := &TenantCluster{
		certsSearcher: config.CertsSearcher,
		logger:        config.Logger,

		certID:          config.CertID,
		tillerNamespace: config.TillerNamespace,
	}

	return g, nil
}

// NewG8sClient returns a generated clientset for the specified tenant cluster.
func (g *TenantCluster) NewG8sClient(ctx context.Context, clusterID, apiDomain string) (versioned.Interface, error) {
	g.logger.LogCtx(ctx, "level", "debug", "message", "creating G8s client for the tenant cluster")

	restConfig, err := g.newRestConfig(ctx, clusterID, apiDomain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	g8sClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	g.logger.LogCtx(ctx, "level", "debug", "message", "created G8s client for the tenant cluster")

	return g8sClient, nil
}

// NewHelmClient returns a Helm client for the specified tenant cluster.
func (g *TenantCluster) NewHelmClient(ctx context.Context, clusterID, apiDomain string) (helmclient.Interface, error) {
	g.logger.LogCtx(ctx, "level", "debug", "message", "creating Helm client for the tenant cluster")

	restConfig, err := g.newRestConfig(ctx, clusterID, apiDomain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	c := helmclient.Config{
		K8sClient: k8sClient,
		Logger:    g.logger,

		RestConfig:      restConfig,
		TillerNamespace: g.tillerNamespace,
	}
	helmClient, err := helmclient.New(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	g.logger.LogCtx(ctx, "level", "debug", "message", "created Helm client for the tenant cluster")

	return helmClient, nil
}

// NewK8sClient returns a Kubernetes clientset for the specified tenant cluster.
func (g *TenantCluster) NewK8sClient(ctx context.Context, clusterID, apiDomain string) (kubernetes.Interface, error) {
	g.logger.LogCtx(ctx, "level", "debug", "message", "creating K8s client for the tenant cluster")

	restConfig, err := g.newRestConfig(ctx, clusterID, apiDomain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	g.logger.LogCtx(ctx, "level", "debug", "message", "created K8s client for the tenant cluster")

	return k8sClient, nil
}

// newRestConfig returns a Kubernetes REST config for the specified tenant
// cluster.
func (g *TenantCluster) newRestConfig(ctx context.Context, clusterID, apiDomain string) (*rest.Config, error) {
	var err error

	g.logger.LogCtx(ctx, "level", "debug", "message", "looking for certificates for the tenant cluster")

	var tls certs.TLS
	{
		tls, err = g.certsSearcher.SearchTLS(clusterID, g.certID)
		if certs.IsTimeout(err) {
			return nil, microerror.Maskf(timeoutError, err.Error())
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	g.logger.LogCtx(ctx, "level", "debug", "message", "found certificates for the tenant cluster")

	var restConfig *rest.Config
	{
		c := k8srestconfig.Config{
			Logger: g.logger,

			Address:   apiDomain,
			InCluster: false,
			TLS: k8srestconfig.ConfigTLS{
				CAData:  tls.CA,
				CrtData: tls.Crt,
				KeyData: tls.Key,
			},
		}
		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return restConfig, nil
}
