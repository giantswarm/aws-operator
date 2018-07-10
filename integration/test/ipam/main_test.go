// +build k8srequired

package ipam

import (
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	e2eclient "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/e2etests/ipam"
	"github.com/giantswarm/e2etests/ipam/provider"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/integration/setup"
)

var (
	c *e2eclient.Client
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
		}

		g, err = framework.NewGuest(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := framework.HostConfig{}

		h, err = framework.NewHost(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c, err = e2eclient.NewClient()
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
	setup.WrapTestMain(c, g, h, m)
}
