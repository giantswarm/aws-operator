// +build k8srequired

package loadtest

import (
	"testing"

	"github.com/giantswarm/helmclient"

	"github.com/giantswarm/e2etests/loadtest"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/integration/setup"
)

var (
	config       setup.Config
	loadTestTest *loadtest.LoadTest
)

func init() {
	var err error

	{
		config, err = setup.NewConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	var controlPlaneHelmClient helmclient.Interface

	{
		c := helmclient.Config{
			Logger:    config.Logger,
			K8sClient: config.Host.K8sClient(),

			RestConfig: config.Host.RestConfig(),
		}

		controlPlaneHelmClient, err = helmclient.New(c)
		if err != nil {
			panic(err.Error())
		}
	}

	var tenantHelmClient helmclient.Interface

	{
		c := helmclient.Config{
			Logger:    config.Logger,
			K8sClient: config.Guest.K8sClient(),

			RestConfig: config.Guest.RestConfig(),
		}

		tenantHelmClient, err = helmclient.New(c)
		if err != nil {
			panic(err.Error())
		}
	}

	var clients *loadtest.Clients

	{
		clients = &loadtest.Clients{
			ControlPlaneHelmClient: controlPlaneHelmClient,
			ControlPlaneK8sClient:  config.Guest.K8sClient(),
			TenantHelmClient:       tenantHelmClient,
			TenantK8sClient:        config.Host.K8sClient(),
		}
	}

	{
		c := loadtest.Config{
			Clients: clients,
			Logger:  config.Logger,

			AuthToken:    env.StormForgerAPIToken(),
			ClusterID:    env.ClusterID(),
			CommonDomain: env.CommonDomain(),
		}

		loadTestTest, err = loadtest.New(c)
		if err != nil {
			panic(err.Error())
		}
	}
}

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	setup.Setup(m, config)
}
