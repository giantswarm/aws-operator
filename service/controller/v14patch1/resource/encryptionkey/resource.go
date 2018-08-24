package encryptionkey

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/v14patch1/encrypter"
)

const (
	// Name is the identifier of the resource.
	Name = "encryptionkeyv14patch1"
)

// Config represents the configuration used to create a new cloudformation resource.
type Config struct {
	// Dependencies.
	Encrypter encrypter.Interface
	Logger    micrologger.Logger

	// Settings.
	InstallationName string
}

// DefaultConfig provides a default configuration to create a new cloudformation
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Encrypter: nil,
		Logger:    nil,
	}
}

// Resource implements the cloudformation resource.
type Resource struct {
	// Dependencies.
	encrypter encrypter.Interface
	logger    micrologger.Logger

	// Settings.
	installationName string
}

// New creates a new configured cloudformation resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.Encrypter == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Encrypter must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	// Settings.
	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	// Settings.
	newService := &Resource{
		// Dependencies.
		encrypter: config.Encrypter,
		logger:    config.Logger,

		// Settings.
		installationName: config.InstallationName,
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}

func toEncryptionKeyState(v interface{}) (encrypter.EncryptionKeyState, error) {
	if v == nil {
		return encrypter.EncryptionKeyState{}, nil
	}

	keyState, ok := v.(encrypter.EncryptionKeyState)
	if !ok {
		return encrypter.EncryptionKeyState{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", keyState, v)
	}

	return keyState, nil
}
