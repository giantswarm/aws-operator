package namespace

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	namespaceToCreate, err := toNamespace(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if namespaceToCreate != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating namespace")

		_, err = r.k8sClient.CoreV1().Namespaces().Create(namespaceToCreate)
		if apierrors.IsAlreadyExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "created namespace")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not create namespace")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentNamespace, err := toNamespace(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredNamespace, err := toNamespace(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var namespaceToCreate *corev1.Namespace
	if currentNamespace == nil {
		namespaceToCreate = desiredNamespace
	}

	return namespaceToCreate, nil
}
