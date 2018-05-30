// Package migration provides an operatorkit resource that migrates awsconfig CRs
// to reference the default credential secret if they do not already.
// It can be safely removed once all awsconfig CRs reference a credential secret.
package migration

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"

	"github.com/giantswarm/aws-operator/service/controller/v12/key"
)

const (
	name = "migrationv12"

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
	customObject = *customObject.DeepCopy()

	oldSpec := *customObject.Spec.DeepCopy()

	if reflect.DeepEqual(providerv1alpha1.AWSConfigSpecAWSHostedZones{}, customObject.Spec.AWS.HostedZones) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "CR is missing hosted zone names")

		apiDomain := customObject.Spec.Cluster.Kubernetes.API.Domain
		zone, err := zoneFromAPIDomain(apiDomain)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("settings all zones to %q", zone))

		customObject.Spec.AWS.HostedZones.API.Name = zone
		customObject.Spec.AWS.HostedZones.Etcd.Name = zone
		customObject.Spec.AWS.HostedZones.Ingress.Name = zone
	}

	if reflect.DeepEqual(customObject.Spec, oldSpec) {
		return nil
	}

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

func zoneFromAPIDomain(apiDomain string) (string, error) {
	parts := strings.Split(apiDomain, ".")
	if len(parts) < 5 {
		return "", microerror.Maskf(malformedDomainError, "API domain must have at least 5 parts, got %d for domain %q", len(parts), apiDomain)
	}

	return strings.Join(parts[3:], "."), nil
}
