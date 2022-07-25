package appsconfig

import (
	"context"
	"reflect"

	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/v12/pkg/project"
	"github.com/giantswarm/aws-operator/v12/service/controller/key"
)

const (
	Name = "appsconfig"

	CiliumAppConfigMapName = "cilium-user-values"
)

type Config struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
}

type Resource struct {
	ctrlClient client.Client
	logger     micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	desiredData := desiredData()
	desiredLabels := map[string]string{
		label.Cluster:   cr.Name,
		label.ManagedBy: project.Name(),
	}

	// Check if ConfigMap exists.
	r.logger.Debugf(ctx, "Checking if ConfigMap for cilium app exists")
	existing := v1.ConfigMap{}
	err = r.ctrlClient.Get(ctx, client.ObjectKey{Name: CiliumAppConfigMapName, Namespace: cr.Name}, &existing)
	if errors.IsNotFound(err) {
		r.logger.Debugf(ctx, "ConfigMap for cilium app does not exist, creating")
		// ConfigMap not found, create it
		cm := v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      CiliumAppConfigMapName,
				Namespace: cr.Name,
				Labels:    desiredLabels,
			},
			Data: desiredData,
		}

		err = r.ctrlClient.Create(ctx, &cm)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "ConfigMap for cilium app was created")
		return nil

	} else if err != nil {
		return microerror.Mask(err)
	}

	// CM exists, check if it's up to date.
	r.logger.Debugf(ctx, "Checking if ConfigMap for cilium is up to date")
	if reflect.DeepEqual(existing.Data, desiredData) && reflect.DeepEqual(existing.Labels, desiredLabels) {
		r.logger.Debugf(ctx, "ConfigMap for cilium app was up to date")
	} else {
		r.logger.Debugf(ctx, "ConfigMap for cilium app needs to be updated")
		existing.Data = desiredData
		existing.Labels = desiredLabels

		err = r.ctrlClient.Update(ctx, &existing)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "ConfigMap for cilium app updated successfully")
	}

	return nil
}

func desiredData() map[string]string {
	values := `ipam:
  mode: kubernetes
defaultPolicies:
  enabled: true
`
	return map[string]string{
		"values": values,
	}
}
