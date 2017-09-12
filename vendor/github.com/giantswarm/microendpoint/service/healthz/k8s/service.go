package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/microendpoint/service/healthz"
)

const (
	// Description describes which functionality this health check implements.
	Description = "Ensure Kubernetes API availability."
	// Name is the identifier of the health check. This can be used for emitting
	// metrics.
	Name = "k8s"
	// SuccessMessage is the message returned in case the health check did not
	// fail.
	SuccessMessage = "all good"
	// Timeout is the time being waited until timing out health check, which
	// renders its result unsuccessful.
	Timeout = 5 * time.Second
)

// Config represents the configuration used to create a healthz service.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	// Settings.
	Timeout time.Duration
}

// DefaultConfig provides a default configuration to create a new healthz service
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		K8sClient: nil,
		Logger:    nil,

		// Settings.
		Timeout: Timeout,
	}
}

// Service implements the healthz service interface.
type Service struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	// Settings.
	timeout time.Duration
}

// New creates a new configured healthz service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
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
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		// Settings.
		timeout: config.Timeout,
	}

	return newService, nil
}

// GetHealthz implements the health check for Kubernetes.
func (s *Service) GetHealthz(ctx context.Context) (healthz.Response, error) {
	failed := false
	message := SuccessMessage
	{
		ch := make(chan string, 1)

		go func() {
			_, err := s.k8sClient.Core().RESTClient().Get().AbsPath("/").DoRaw()
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
