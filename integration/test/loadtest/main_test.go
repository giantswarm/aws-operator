// +build k8srequired

package loadtest

import (
	"testing"

	"github.com/giantswarm/e2etests/loadtest"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/integration/setup"
)

var (
	config       setup.Config
	loadTestTest *loadtest.LoadTest
)

func init() {
	var err error

	{
		config, err = setup.NewConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := loadtest.Config{
			GuestFramework: config.Guest,
			Logger:         config.Logger,

			AuthToken:    env.StormForgerAPIToken(),
			ClusterID:    env.ClusterID(),
			CommonDomain: env.CommonDomain(),
		}

		loadTestTest, err = loadtest.New(c)
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
