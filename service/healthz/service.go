package healthz

import (
	"context"

	"github.com/aws/aws-sdk-go/service/iam"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
)

const (
	// awsRegion is required even though the IAM API is global.
	awsRegion string = "eu-central-1"
)

// Config represents the configuration used to create a healthz service.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger

	// Config
	AwsConfig awsutil.Config
}

// DefaultConfig provides a default configuration to create a new healthz
// service by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger: nil,

		// Settings.
		AwsConfig: awsutil.Config{},
	}
}

// New creates a new configured healthz service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}

	// Settings.
	var emptyAwsConfig awsutil.Config
	if config.AwsConfig == emptyAwsConfig {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.AwsConfig must not be empty")
	}

	newService := &Service{
		// Dependencies.
		logger: config.Logger,

		// Settings.
		awsConfig: config.AwsConfig,
	}

	return newService, nil
}

// Service implements the healthz service interface.
type Service struct {
	// Dependencies.
	logger micrologger.Logger

	// Settings.
	awsConfig awsutil.Config
}

// Check implements the health check which gets the current user to check
// we can authenticate.
// TODO Check the user has access to the correct set of AWS services.
func (s *Service) Check(ctx context.Context, request Request) (*Response, error) {
	// Set the region for the API client.
	s.awsConfig.Region = awsRegion
	clients := awsutil.NewClients(s.awsConfig)

	// Get the current user.
	_, err := clients.IAM.GetUser(&iam.GetUserInput{})
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return DefaultResponse(), nil
}
