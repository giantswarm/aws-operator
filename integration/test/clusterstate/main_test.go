// +build k8srequired

package clusterstate

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	e2eclient "github.com/giantswarm/e2eclients/aws"
	e2esetup "github.com/giantswarm/e2esetup/aws"
	"github.com/giantswarm/e2esetup/aws/env"
	"github.com/giantswarm/e2etests/clusterstate"
	"github.com/giantswarm/e2etests/clusterstate/provider"
	"github.com/giantswarm/micrologger"
)

var (
	c  *e2eclient.Client
	cs *clusterstate.ClusterState
	g  *framework.Guest
	h  *framework.Host
	l  micrologger.Logger
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

			ClusterID: env.ClusterID(),
		}

		p, err = provider.NewAWS(ac)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		cc := clusterstate.Config{
			GuestFramework: g,
			Logger:         l,
			Provider:       p,
		}

		cs, err = clusterstate.New(cc)
		if err != nil {
			panic(err.Error())
		}
	}
}

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	ctx := context.Background()

	{
		c := e2esetup.Config{
			AWSClient: c,
			Guest:     g,
			Host:      h,
		}

		err := e2esetup.Setup(ctx, m, c)
		if err != nil {
			l.LogCtx(ctx, "level", "error", "message", "e2e test failed", "stack", fmt.Sprintf("%#v\n", err))
			os.Exit(1)
		}
	}

	os.Exit(0)
}
