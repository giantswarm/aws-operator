// +build k8srequired

package update

import (
	"testing"
	"time"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/e2etests/update"
	"github.com/giantswarm/e2etests/update/provider"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/integration/setup"
)

var (
	c *aws.Client
	g *framework.Guest
	h *framework.Host
	u *update.Update
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

	var p *provider.AWS
	{
		c := provider.AWSConfig{
			HostFramework: h,
			Logger:        l,

			ClusterID:   env.ClusterID(),
			GithubToken: env.GithubToken(),
		}

		p, err = provider.NewAWS(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := update.Config{
			Logger:   l,
			Provider: p,

			MaxWait: 90 * time.Minute,
		}

		u, err = update.New(c)
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
}

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	setup.WrapTestMain(c, g, h, m)
}
