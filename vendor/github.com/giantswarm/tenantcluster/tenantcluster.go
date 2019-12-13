package tenantcluster

import (
	"context"

	"github.com/giantswarm/certs"
	"github.com/giantswarm/k8sclient/k8srestconfig"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/rest"
)

// Config represents the configuration used to create a new tenant cluster
// service.
type Config struct {
	CertsSearcher certs.Interface
	Logger        micrologger.Logger

	CertID certs.Cert
}

// TenantCluster provides functionality for connecting to tenant clusters.
type TenantCluster struct {
	certsSearcher certs.Interface
	logger        micrologger.Logger

	certID certs.Cert
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

	t := &TenantCluster{
		certsSearcher: config.CertsSearcher,
		logger:        config.Logger,

		certID: config.CertID,
	}

	return t, nil
}

func (t *TenantCluster) NewRestConfig(ctx context.Context, clusterID, apiDomain string) (*rest.Config, error) {
	var err error

	t.logger.LogCtx(ctx, "level", "debug", "message", "looking for certificates for the tenant cluster")

	var tls certs.TLS
	{
		tls, err = t.certsSearcher.SearchTLS(clusterID, t.certID)
		if certs.IsTimeout(err) {
			return nil, microerror.Maskf(timeoutError, err.Error())
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	t.logger.LogCtx(ctx, "level", "debug", "message", "found certificates for the tenant cluster")

	var restConfig *rest.Config
	{
		c := k8srestconfig.Config{
			Logger: t.logger,

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
