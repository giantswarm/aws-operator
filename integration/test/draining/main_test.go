// +build k8srequired

package scaling

import (
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"

	"github.com/giantswarm/aws-operator/integration/client"
	"github.com/giantswarm/aws-operator/integration/setup"
)

var (
	g *framework.Guest
	h *framework.Host
	c *client.AWS
)

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	var err error

	g, err = framework.NewGuest()
	if err != nil {
		panic(err.Error())
	}
	h, err = framework.NewHost(framework.HostConfig{})
	if err != nil {
		panic(err.Error())
	}

	c = client.NewAWS()

	setup.WrapTestMain(c, g, h, m)
}
