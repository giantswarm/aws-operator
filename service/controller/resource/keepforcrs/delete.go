package keepforcrs

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/finalizerskeptcontext"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var list *unstructured.UnstructuredList
	{
		gvk, err := apiutil.GVKForObject(r.newObjFunc(), r.k8sClient.Scheme())
		if err != nil {
			return microerror.Mask(err)
		}
		gvk.Kind += "List"

		l := &unstructured.UnstructuredList{}
		l.SetGroupVersionKind(gvk)

		list = l
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding objects of type %T for tenant cluster %#q", r.newObjFunc(), key.ClusterID(cr)))

		err = r.k8sClient.CtrlClient().List(
			ctx,
			list,
			client.InNamespace(cr.GetNamespace()),
			client.MatchingLabels{label.Cluster: key.ClusterID(cr)},
		)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d object(s) of type %T for tenant cluster %#q", len(list.Items), r.newObjFunc(), key.ClusterID(cr)))
	}

	if len(list.Items) != 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")
		finalizerskeptcontext.SetKept(ctx)
	}

	return nil
}
