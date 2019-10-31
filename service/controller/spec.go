package controller

import "github.com/giantswarm/micrologger"

type encrypterConfigGetter interface {
	GetEncrypterBackend() string
	GetInstallationName() string
	GetLogger() micrologger.Logger
	GetVaultAddress() string
}
