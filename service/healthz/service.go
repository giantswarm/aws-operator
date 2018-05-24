package healthz

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/iam"
	healthzservice "github.com/giantswarm/microendpoint/service/healthz"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
)

const (
	// AWSRegion is required even though the IAM API is global.
	AWSRegion string = "eu-central-1"

	// Description describes which functionality this health check implements.
	Description = "Ensure AWS API availability."
	// Name is the identifier of the health check. This can be used for emitting
	// metrics.
	Name = "aws"
	// SuccessMessage is the message returned in case the health check did not
	// fail.
	SuccessMessage = "all good"
	// Timeout is the time being waited until timing out health check, which
	// renders its result unsuccessful.
	Timeout = 5 * time.Second
)

// Config represents the configuration used to create a healthz service.
type Config struct {
	Logger micrologger.Logger

	AwsConfig awsutil.Config
}

// New creates a new configured healthz service.
func New(config Config) (*Service, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	var emptyAwsConfig awsutil.Config
	if config.AwsConfig == emptyAwsConfig {
		return nil, microerror.Maskf(invalidConfigError, "config.AwsConfig must not be empty")
	}

	newService := &Service{
		logger: config.Logger,

		awsConfig: config.AwsConfig,
	}

	return newService, nil
}

// Service implements the healthz service interface.
type Service struct {
	logger micrologger.Logger

	awsConfig awsutil.Config
}

// GetHealthz implements the health check for AWS. It does this by calling the
// IAM API to get the current user. This checks that we can connect to the API
// and the credentials are correct.
func (s *Service) GetHealthz(ctx context.Context) (healthzservice.Response, error) {
	failed := false
	message := SuccessMessage
	{
		ch := make(chan string, 1)

		go func() {
			// Set the region for the API client.
			if s.awsConfig.Region == "" {
				s.awsConfig.Region = AWSRegion
			}
			clients := awsutil.NewClients(s.awsConfig)

			// Get the current user.
			_, err := clients.IAM.GetUser(&iam.GetUserInput{})
			if err != nil {
				ch <- err.Error()
				return
			}

			ch <- ""
		}()

		select {
		case m := <-ch:
			if m != "" {
				failed = true
				message = m
			}
		case <-time.After(Timeout):
			failed = true
			message = fmt.Sprintf("timed out after %s", Timeout)
		}
	}

	response := healthzservice.Response{
		Description: Description,
		Failed:      failed,
		Message:     message,
		Name:        Name,
	}

	return response, nil
}
