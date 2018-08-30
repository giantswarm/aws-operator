// +build k8srequired

package scaling

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2eclients/aws"
	e2esetup "github.com/giantswarm/e2esetup/aws"
	"github.com/giantswarm/e2esetup/aws/env"
	"github.com/giantswarm/e2etests/scaling"
	"github.com/giantswarm/e2etests/scaling/provider"
	"github.com/giantswarm/micrologger"
)

var (
	c *aws.Client
	g *framework.Guest
	h *framework.Host
	l micrologger.Logger
	s *scaling.Scaling
)

func init() {
	var err error

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
			GuestFramework: g,
			HostFramework:  h,
			Logger:         l,

			ClusterID: env.ClusterID(),
		}

		p, err = provider.NewAWS(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := scaling.Config{
			Logger:   l,
			Provider: p,
		}

		s, err = scaling.New(c)
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
	ctx := context.Background()

	var err error
	var v int

	{
		c := e2esetup.Config{
			AWSClient: c,
			Guest:     g,
			Host:      h,
		}

		err = e2esetup.Setup(ctx, m, c)
		if err != nil {
			l.Log(fmt.Sprintf("%#v\n", err))
			v = 1
		}
	}

	os.Exit(v)
}
