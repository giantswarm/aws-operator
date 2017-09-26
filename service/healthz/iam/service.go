package iam

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/giantswarm/microendpoint/service/healthz"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	// Description describes which functionality this health check implements.
	Description = "Ensure IAM API availability."
	// Name is the identifier of the health check. This can be used for emitting
	// metrics.
	Name = "iam"
	// SuccessMessage is the message returned in case the health check did not
	// fail.
	SuccessMessage = "all good"
	// Timeout is the time being waited until timing out health check, which
	// renders its result unsuccessful.
	Timeout = 5 * time.Second
)

const (
	// awsRegion is required even though the IAM API is global.
	awsRegion string = "eu-central-1"
)

// Config represents the configuration used to create a healthz service.
type Config struct {
	// Dependencies.
	IAMClient *iam.IAM
	Logger    micrologger.Logger

	// Settings.
	Timeout time.Duration
}

// DefaultConfig provides a default configuration to create a new healthz
// service by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		IAMClient: nil,
		Logger:    nil,

		// Settings.
		Timeout: Timeout,
	}
}

// Service implements the healthz service interface.
type Service struct {
	// Dependencies.
	iamClient *iam.IAM
	logger    micrologger.Logger

	// Settings.
	timeout time.Duration
}

// New creates a new configured healthz service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.IAMClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.IAMClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	// Settings.
	if config.Timeout.Seconds() == 0 {
		return nil, microerror.Maskf(invalidConfigError, "config.Timout must not be empty")
	}

	newService := &Service{
		// Dependencies.
		iamClient: config.IAMClient,
		logger:    config.Logger,

		// Settings.
		timeout: config.Timeout,
	}

	return newService, nil
}

// GetHealthz implements the health check for AWS IAM.
func (s *Service) GetHealthz(ctx context.Context) (healthz.Response, error) {
	failed := false
	message := SuccessMessage
	{
		ch := make(chan string, 1)

		go func() {
			_, err := s.iamClient.GetUser(&iam.GetUserInput{})
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
		case <-time.After(s.timeout):
			failed = true
			message = fmt.Sprintf("timed out after %s", s.timeout)
		}
	}

	response := healthz.Response{
		Description: Description,
		Failed:      failed,
		Message:     message,
		Name:        Name,
	}

	return response, nil
}
