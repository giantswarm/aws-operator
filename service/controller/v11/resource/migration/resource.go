// Package migration provides an operatorkit resource that migrates awsconfig CRs
// to reference the default credential secret if they do not already.
// It can be safely removed once all awsconfig CRs reference a credential secret.
package migration

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/aws-operator/service/controller/v11/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "migrationv11"

	AWSConfigNamespace               = "default"
	CredentialSecretDefaultNamespace = "giantswarm"
	CredentialSecretDefaultName      = "credential-default"
)

type Config struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger
}

type Resource struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	newResource := &Resource{
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// If the default credential secret is not referenced,
	// update the CR, and then cancel reconciliation.
	credentialSecret := customObject.Spec.AWS.CredentialSecret
	if credentialSecret.Namespace == "" || credentialSecret.Name == "" {
		r.logger.Log("level", "debug", "message", "config is missing default credential, migrating")

		customObject.Spec.AWS.CredentialSecret.Namespace = CredentialSecretDefaultNamespace
		customObject.Spec.AWS.CredentialSecret.Name = CredentialSecretDefaultName

		_, err := r.g8sClient.ProviderV1alpha1().AWSConfigs(AWSConfigNamespace).Update(&customObject)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
