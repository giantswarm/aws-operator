package namespace

import (
	"context"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/aws-operator/service/awsconfig/v10/key"
)

// EnsureCreated creates a Kubernetes namespace in the host cluster for the
// guest cluster.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "creating Kubernetes namespace")

	namespaceToCreate := getNamespace(customObject)

	_, err = r.k8sClient.CoreV1().Namespaces().Create(namespaceToCreate)
	if apierrors.IsAlreadyExists(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating Kubernetes namespace: already created")
		return nil

	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "creating Kubernetes namespace: created")

	return nil
}
