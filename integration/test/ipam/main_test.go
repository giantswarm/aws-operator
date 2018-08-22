// +build k8srequired

package ipam

import (
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	aws "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/e2etests/ipam"
	"github.com/giantswarm/e2etests/ipam/provider"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/integration/setup"
)

var (
	c *aws.Client
	g *framework.Guest
	i *ipam.IPAM
	h *framework.Host
)

func init() {
	var err error

	var l micrologger.Logger
	{
		c := micrologger.Config{}

		l, err = micrologger.New(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := framework.GuestConfig{
			Logger: l,

			ClusterID:    env.ClusterID(),
			CommonDomain: env.CommonDomain(),
		}

		g, err = framework.NewGuest(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := framework.HostConfig{
			Logger: l,

			ClusterID:  env.ClusterID(),
			VaultToken: env.VaultToken(),
		}

		h, err = framework.NewHost(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c, err = aws.NewClient()
		if err != nil {
			panic(err.Error())
		}
	}

	var p *provider.AWS
	{

		ac := provider.AWSConfig{
			AWSClient:     c,
			HostFramework: h,
			Logger:        l,

			ChartValuesConfig: provider.ChartValuesConfig{
				ClusterName: env.ClusterID(),
			},
		}

		p, err = provider.NewAWS(ac)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		ic := ipam.Config{
			Logger:   l,
			Provider: p,

			CommonDomain: env.CommonDomain(),
		}

		i, err = ipam.New(ic)
		if err != nil {
			panic(err.Error())
		}
	}
}

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	{
		c := setup.Config{
			AWSClient: c,
			Guest:     g,
			Host:      h,
		}

		setup.Setup(m, c)
	}
}
