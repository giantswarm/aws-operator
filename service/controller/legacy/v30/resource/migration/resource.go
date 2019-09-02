// Package migration provides an operatorkit resource that migrates awsconfig CRs
// to reference the default credential secret if they do not already.
// It can be safely removed once all awsconfig CRs reference a credential secret.
//
// Latest changes:
//
//	* v24: Added migration code to fill spec.Cluster.Scaling.{Min,Max} values
//	       when they are zero. When all tenant clusters are migrated to at least v24,
//	       migrations in this file can be removed.
//
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/legacy/v30/key"
)

const (
	name = "migrationv30"

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
	var customObject providerv1alpha1.AWSConfig
	var oldSpec providerv1alpha1.AWSConfigSpec
	{
		o, err := key.ToCustomObject(obj)
		if err != nil {
			return microerror.Mask(err)
		}
		// We have to always fetch the latest version of the resource in order to
		// update it below using the latest resource version.
		m, err := r.g8sClient.ProviderV1alpha1().AWSConfigs(o.GetNamespace()).Get(o.GetName(), metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		customObject = *m.DeepCopy()
		oldSpec = *m.Spec.DeepCopy()

		err = r.migrateSpec(ctx, &customObject.Spec)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating CR with version %#q", customObject.ResourceVersion))

		if reflect.DeepEqual(customObject.Spec, oldSpec) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("CR with version %#q is already up to date", customObject.ResourceVersion))
		} else {
			_, err := r.g8sClient.ProviderV1alpha1().AWSConfigs(customObject.GetNamespace()).Update(&customObject)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated CR with version %#q", customObject.ResourceVersion))

			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
			reconciliationcanceledcontext.SetCanceled(ctx)
		}
	}

	return nil
}

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}

func (r *Resource) migrateSpec(ctx context.Context, spec *providerv1alpha1.AWSConfigSpec) error {
	if spec.AWS.CredentialSecret.Name == "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "CR is missing credential secret")

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting CR credential secret to %#q in namespace %#q", credentialSecretDefaultName, credentialSecretDefaultNamespace))

		spec.AWS.CredentialSecret.Namespace = credentialSecretDefaultNamespace
		spec.AWS.CredentialSecret.Name = credentialSecretDefaultName
	}

	if reflect.DeepEqual(providerv1alpha1.AWSConfigSpecAWSHostedZones{}, spec.AWS.HostedZones) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "CR is missing hosted zone names")

		apiDomain := spec.Cluster.Kubernetes.API.Domain
		zone, err := zoneFromAPIDomain(apiDomain)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting CR hosted zone names to %#q", zone))

		spec.AWS.HostedZones.API.Name = zone
		spec.AWS.HostedZones.Etcd.Name = zone
		spec.AWS.HostedZones.Ingress.Name = zone
	}

	if spec.Cluster.Scaling.Min == 0 && spec.Cluster.Scaling.Max == 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "CR is missing cluster scaling configuration")

		nWorkers := len(spec.AWS.Workers)

		spec.Cluster.Scaling.Min = nWorkers
		spec.Cluster.Scaling.Max = nWorkers
	}

	return nil
}

func zoneFromAPIDomain(apiDomain string) (string, error) {
	parts := strings.Split(apiDomain, ".")
	if len(parts) < 5 {
		return "", microerror.Maskf(malformedDomainError, "API domain must have at least 5 parts, got %d for domain %q", len(parts), apiDomain)
	}

	return strings.Join(parts[3:], "."), nil
}
