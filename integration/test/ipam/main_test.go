// +build k8srequired

package ipam

import (
	"testing"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/integration/setup"
	"github.com/giantswarm/e2etests/ipam"
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

		// We disable the default tenant cluster here, because the IPAM tests runs
		// multiple customized tenant clusters.
		config.UseDefaultTenant = false
	}

	var p *Provider
	{
		c := ProviderConfig{
			AWSClient: config.AWSClient,
			Host:      config.Host,
			Logger:    config.Logger,
			Release:   config.Release,
		}

		p, err = NewProvider(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := ipam.Config{
			Logger:   config.Logger,
			Provider: p,

			ClusterID: env.ClusterID(),
		}

		ipamTest, err = ipam.New(c)
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
