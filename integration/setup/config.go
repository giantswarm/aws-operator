package setup

import (
	"github.com/giantswarm/e2e-harness/pkg/framework"
	awsclient "github.com/giantswarm/e2eclients/aws"
)

type Config struct {
	AWSClient *awsclient.Client
	Guest     *framework.Guest
	Host      *framework.Host
}
