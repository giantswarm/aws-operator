// Package migration provides an operatorkit resource that migrates awsconfig CRs
// to reference the default credential secret if they do not already.
// It can be safely removed once all awsconfig CRs reference a credential secret.
package migration

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/aws-operator/service/controller/v12patch1/key"
)

const (
	name = "migrationv12patch1"

	awsConfigNamespace               = "default"
	credentialSecretDefaultNamespace = "giantswarm"
	credentialSecretDefaultName      = "credential-default"
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

	r := &Resource{
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return name
}

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// If credential secret is set we have nothing to migrate.
	if customObject.Spec.AWS.CredentialSecret.Name != "" {
		return nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "CR is missing credential, setting the default")

	customObject.Spec.AWS.CredentialSecret.Namespace = credentialSecretDefaultNamespace
	customObject.Spec.AWS.CredentialSecret.Name = credentialSecretDefaultName

	r.logger.LogCtx(ctx, "level", "debug", "message", "updating CR")

	_, err = r.g8sClient.ProviderV1alpha1().AWSConfigs(awsConfigNamespace).Update(&customObject)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "CR updated, canceling reconciliation")
	reconciliationcanceledcontext.SetCanceled(ctx)

	return nil
}

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
