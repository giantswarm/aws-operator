// +build k8srequired

package all

import (
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"

	"github.com/giantswarm/aws-operator/integration/client"
	"github.com/giantswarm/aws-operator/integration/setup"
)

var (
	f *framework.Framework
	c *client.AWS
)

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	var err error

	f, err = framework.New()
	if err != nil {
		panic(err.Error())
	}

	c = client.NewAWS()

	setup.WrapTestMain(c, f, m)
}
