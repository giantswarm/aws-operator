package controller

import "github.com/giantswarm/micrologger"

type encrypterConfigGetter interface {
	GetInstallationName() string
	GetLogger() micrologger.Logger
	GetVaultAddress() string
}
