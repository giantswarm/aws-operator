package namespace

import (
	"context"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/awsconfig/v10/key"
)

// EnsureDeleted deletes the Kubernetes namespace in the host cluster for the
// guest cluster.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "deleting Kubernetes namespace")

	namespaceToDelete := getNamespace(customObject)

	err = r.k8sClient.CoreV1().Namespaces().Delete(namespaceToDelete.Name, &apismetav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting Kubernetes namespace: already deleted")
		return nil

	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "deleting Kubernetes namespace: deleted")

	return nil
}
