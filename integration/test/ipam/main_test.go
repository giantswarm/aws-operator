// +build k8srequired

package ipam

import (
	"testing"

	"github.com/giantswarm/e2etests/ipam"
	"github.com/giantswarm/e2etests/ipam/provider"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/integration/setup"
)

var (
	config   setup.Config
	ipamTest *ipam.IPAM
)

func init() {
	var err error

	{
		config, err = setup.NewConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	var p *provider.AWS
	{

		ac := provider.AWSConfig{
			AWSClient:     config.AWSClient,
			HostFramework: config.Host,
			Logger:        config.Logger,

			ChartValuesConfig: provider.ChartValuesConfig{
				AWSRouteTable0: env.ClusterID() + "_0",
				AWSRouteTable1: env.ClusterID() + "_1",
				ClusterName:    env.ClusterID(),
			},
		}

		p, err = provider.NewAWS(ac)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		ic := ipam.Config{
			HostFramework: config.Host,
			Logger:        config.Logger,
			Provider:      p,

			CommonDomain:    env.CommonDomain(),
			HostClusterName: env.ClusterID(),
		}

		ipamTest, err = ipam.New(ic)
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
