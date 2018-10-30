// +build k8srequired

package scaling

import (
	"testing"

	"github.com/giantswarm/e2etests/scaling"
	"github.com/giantswarm/e2etests/scaling/provider"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/integration/setup"
)

var (
	config      setup.Config
	scalingTest *scaling.Scaling
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
		c := provider.AWSConfig{
			GuestFramework: config.Guest,
			HostFramework:  config.Host,
			Logger:         config.Logger,

			ClusterID: env.ClusterID(),
		}

		p, err = provider.NewAWS(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := scaling.Config{
			Logger:   config.Logger,
			Provider: p,
		}

		scalingTest, err = scaling.New(c)
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
