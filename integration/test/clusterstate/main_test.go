package clusterstate

import (
	"testing"

	"github.com/giantswarm/e2etests/clusterstate"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/integration/setup"
)

var (
	config           setup.Config
	clusterStateTest *clusterstate.ClusterState
)

func init() {
	var err error

	{
		config, err = setup.NewConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	var p *Provider
	{
		c := ProviderConfig{
			AWSClient: config.AWSClient,
			G8sClient: config.K8sClients.G8sClient(),
			Logger:    config.Logger,

			ClusterID: env.ClusterID(),
		}

		p, err = NewProvider(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := clusterstate.Config{
			LegacyFramework: config.Guest,
			Logger:          config.Logger,
			Provider:        p,
		}

		clusterStateTest, err = clusterstate.New(c)
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
