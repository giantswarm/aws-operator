package statusresource

import (
	"context"
	"fmt"
	"time"

	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", "patching CR status")

	// We process the status updates within its own backoff here to gurantee its
	// execution independent of any eventual retries via the retry resource. It
	// might happen that the reconciled object is not the latest version so any
	// patch would fail. In case the patch fails we retry until we succeed. The
	// steps of the backoff operation are as follows.
	//
	//     Fetch latest version of runtime object.
	//     Compute patches for runtime object.
	//     Apply computed list of patches.
	//
	// In case there are no patches we do not need to do anything. So we prevent
	// unnecessary API calls.
	var modified bool
	{
		o := func() error {
			accessor, err := meta.Accessor(obj)
			if err != nil {
				return microerror.Mask(err)
			}

			newObj, err := r.restClient.Get().AbsPath(accessor.GetSelfLink()).Do().Get()
			if errors.IsNotFound(err) {
				return backoff.Permanent(microerror.Mask(notFoundError))
			} else if err != nil {
				return microerror.Mask(err)
			}

			newAccessor, err := meta.Accessor(newObj)
			if err != nil {
				return microerror.Mask(err)
			}

			patches, err := r.computeDeleteEventPatches(ctx, newObj)
			if err != nil {
				return microerror.Mask(err)
			}

			if len(patches) > 0 {
				err := r.applyPatches(ctx, newAccessor, patches)
				if err != nil {
					return microerror.Mask(err)
				}

				modified = true
			}

			return nil
		}
		b := r.backOffFactory()
		n := func(err error, d time.Duration) {
			r.logger.LogCtx(ctx, "level", "warning", "message", "retrying status patching due to error", "stack", fmt.Sprintf("%#v", err))
		}

		err := backoff.RetryNotify(o, b, n)
		if IsNotFound(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	if modified {
		r.logger.LogCtx(ctx, "level", "debug", "message", "patched CR status")

		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)

		r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")
		finalizerskeptcontext.SetKept(ctx)
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not patch CR status")
	}

	return nil
}

func (r *Resource) computeDeleteEventPatches(ctx context.Context, obj interface{}) ([]Patch, error) {
	clusterStatus, err := r.clusterStatusFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var patches []Patch

	// In case a CR might not have a status at all, we cannot work with it below.
	// We have to initialize it upfront to be safe. Note that we only initialize
	// fields that are managed by the statusresource library implementation. There
	// might be other properties managed by external authorities who have to
	// manage their own initialization.
	patches = ensureDefaultPatches(clusterStatus, patches)

	// Ensure the cluster is set into a deleting status on the delete event.
	{
		notDeleting := !clusterStatus.HasDeletingCondition()

		if notDeleting {
			patches = append(patches, Patch{
				Op:    "replace",
				Path:  "/status/cluster/conditions",
				Value: clusterStatus.WithDeletingCondition(),
			})

			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("setting %#q status condition", providerv1alpha1.StatusClusterTypeDeleting))
		}
	}

	return patches, nil
}
