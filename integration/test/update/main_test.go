// +build k8srequired

package update

import (
	"testing"
	"time"

	"github.com/giantswarm/e2etests/update"
	"github.com/giantswarm/e2etests/update/provider"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/integration/setup"
)

var (
	config     setup.Config
	updateTest *update.Update
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
			HostFramework: config.Host,
			Logger:        config.Logger,

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
			Logger:   config.Logger,
			Provider: p,

			MaxWait: 90 * time.Minute,
		}

		updateTest, err = update.New(c)
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
