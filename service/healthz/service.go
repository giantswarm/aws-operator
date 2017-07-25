package healthz

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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

		// Internals.
		bootOnce: sync.Once{},
	}

	return newService, nil
}

// Service implements the healthz service interface.
type Service struct {
	// Dependencies.
	logger micrologger.Logger

	// Settings.
	awsConfig awsutil.Config

	// Internals.
	bootOnce sync.Once
}

// Check implements the health check which gets the current user and checks
// it belongs to at least one group.
// TODO Check the user is in the aws-operator group to make sure the
// permissions are correct. This group needs to be created.
func (s *Service) Check(ctx context.Context, request Request) (*Response, error) {
	start := time.Now()
	defer func() {
		healthCheckRequestTime.Set(float64(time.Since(start) / time.Millisecond))
	}()

	// Set the region for the API client.
	s.awsConfig.Region = awsRegion
	clients := awsutil.NewClients(s.awsConfig)

	// Get the current user.
	user, err := clients.IAM.GetUser(&iam.GetUserInput{})
	if err != nil {
		healthCheckRequests.WithLabelValues(PrometheusFailedLabel).Inc()
		return nil, microerror.MaskAny(err)
	}

	// Get the groups the current user belongs to.
	userName := *user.User.UserName
	groups, err := clients.IAM.ListGroupsForUser(&iam.ListGroupsForUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		healthCheckRequests.WithLabelValues(PrometheusFailedLabel).Inc()
		return nil, microerror.MaskAny(err)
	}

	// Check the user belongs to at least one group.
	// TODO Check the user belongs to the aws-operator group.
	if len(groups.Groups) == 0 {
		s.logger.Log("info", fmt.Sprintf("Healthcheck failed. User '%s' belongs to 0 groups.", userName))

		healthCheckRequests.WithLabelValues(PrometheusFailedLabel).Inc()
		return nil, microerror.MaskAny(err)
	}

	healthCheckRequests.WithLabelValues(PrometheusSuccessfulLabel).Inc()

	return DefaultResponse(), nil
}
