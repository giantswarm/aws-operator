package kmskey

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/v6/key"
)

const (
	// Name is the identifier of the resource.
	Name = "kmskeyv6"
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

func toAlias(keyID string) string {
	return fmt.Sprintf("alias/%s", keyID)
}

func getKMSTags(customObject v1alpha1.AWSConfig) []*kms.Tag {
	clusterTags := key.ClusterTags(customObject)
	kmsTags := []*kms.Tag{}

	for k, v := range clusterTags {
		tag := &kms.Tag{
			TagKey:   aws.String(k),
			TagValue: aws.String(v),
		}

		kmsTags = append(kmsTags, tag)
	}

	return kmsTags
}
