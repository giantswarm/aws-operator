package vault

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	vaultapi "github.com/hashicorp/vault/api"

	"github.com/giantswarm/microendpoint/service/healthz"
)

const (
	// Description describes which functionality this health check implements.
	Description = "Ensure Vault API availability."
	// Name is the identifier of the health check. This can be used for emitting
	// metrics.
	Name = "vault"
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
	Logger      micrologger.Logger
	VaultClient *vaultapi.Client

	// Settings.
	Timeout time.Duration
}

// DefaultConfig provides a default configuration to create a new healthz
// service by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:      nil,
		VaultClient: nil,

		// Settings.
		Timeout: Timeout,
	}
}

// Service implements the healthz service interface.
type Service struct {
	// Dependencies.
	logger      micrologger.Logger
	vaultClient *vaultapi.Client

	// Settings.
	timeout time.Duration
}

// New creates a new configured healthz service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}
	if config.VaultClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "vault client must not be empty")
	}

	// Settings.
	if config.Timeout.Seconds() == 0 {
		return nil, microerror.Maskf(invalidConfigError, "config.Timout must not be empty")
	}

	newService := &Service{
		// Dependencies.
		logger:      config.Logger,
		vaultClient: config.VaultClient,

		// Settings.
		timeout: config.Timeout,
	}

	return newService, nil
}

// GetHealthz implements the health check for Vault. It does this by listing the
// mounts for the sys backend. This checks that the we can connect to the Vault
// API and that the Vault token is valid.
func (s *Service) GetHealthz(ctx context.Context) (healthz.Response, error) {
	failed := false
	message := SuccessMessage
	{
		ch := make(chan string, 1)

		go func() {
			_, err := s.vaultClient.Sys().ListMounts()
			if err != nil {
				if strings.Contains(err.Error(), "permission denied") {
					setVaultPermissionDenied()
				} else {
					setVaultUnknownError()
				}

				ch <- err.Error()
				return
			}

			setVaultOK()
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
