package setup

import (
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2e-harness/pkg/release"
	awsclient "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/micrologger"
)

type Config struct {
	AWSClient *awsclient.Client
	Guest     *framework.Guest
	Host      *framework.Host
	Logger    micrologger.Logger
	Release   *release.Release
}
