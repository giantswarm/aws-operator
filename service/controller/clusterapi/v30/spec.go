package v30

import "github.com/giantswarm/micrologger"

type EncrypterConfigGetter interface {
	GetEncrypterBackend() string
	GetInstallationName() string
	GetLogger() micrologger.Logger
	GetVaultAddress() string
}
