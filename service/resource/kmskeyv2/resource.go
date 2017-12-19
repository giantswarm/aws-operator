package kmskeyv2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
)

const (
	// Name is the identifier of the resource.
	Name = "kmskeyv2"
)

// Config represents the configuration used to create a new cloudformation resource.
type Config struct {
	// Dependencies.
	Clients Clients
	Logger  micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new cloudformation
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Clients: Clients{},
		Logger:  nil,
	}
}

// Resource implements the cloudformation resource.
type Resource struct {
	// Dependencies.
	awsClients Clients
	logger     micrologger.Logger
}

// New creates a new configured cloudformation resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Resource{
		// Dependencies.
		awsClients: config.Clients,
		logger: config.Logger.With(
			"resource", Name,
		),
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, deleteChange interface{}) error {
	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	return &framework.Patch{}, nil
}

func toKMSKeyState(v interface{}) (KMSKeyState, error) {
	if v == nil {
		return KMSKeyState{}, nil
	}

	kmsKey, ok := v.(KMSKeyState)
	if !ok {
		return KMSKeyState{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", kmsKey, v)
	}

	return kmsKey, nil
}

func toCreateKeyInput(v interface{}) (*kms.CreateKeyInput, error) {
	if v == nil {
		return nil, nil
	}

	createKeyInput, ok := v.(*kms.CreateKeyInput)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", createKeyInput, v)
	}

	return createKeyInput, nil
}

func toDeleteAliasInput(v interface{}) (kms.DeleteAliasInput, error) {
	if v == nil {
		return kms.DeleteAliasInput{}, nil
	}

	deleteAliasInput, ok := v.(kms.DeleteAliasInput)
	if !ok {
		return kms.DeleteAliasInput{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", deleteAliasInput, v)
	}

	return deleteAliasInput, nil
}

func toAlias(keyID string) string {
	return fmt.Sprintf("alias/%s", keyID)
}
