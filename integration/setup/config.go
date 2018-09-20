package setup

import (
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2e-harness/pkg/framework/resource"
	awsclient "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/micrologger"
)

type Config struct {
	AWSClient *awsclient.Client
	Guest     *framework.Guest
	Host      *framework.Host
	Resource  *resource.Resource
	Logger    micrologger.Logger
}
