package restrictawsnodedaemonset

import (
	"context"

	"github.com/giantswarm/microerror"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/v13/pkg/label"
	"github.com/giantswarm/aws-operator/v13/pkg/project"
	"github.com/giantswarm/aws-operator/v13/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v13/service/controller/key"
)

const (
	dsNamespace     = "kube-system"
	awsNodeDsName   = "aws-node"
	KubeProxyDsName = "kube-proxy"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	var err error

	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	hasCilium, err := key.HasCilium(&cr)
	if err != nil {
		return microerror.Mask(err)
	}

	if !hasCilium {
		r.logger.Debugf(ctx, "This cluster has no Cilium.")
		r.logger.Debugf(ctx, "canceling resource")

		return nil
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if cc.Client.TenantCluster.K8s == nil {
		r.logger.Debugf(ctx, "kubernetes clients are not available in controller context yet")
		r.logger.Debugf(ctx, "canceling resource")

		return nil
	}

	ctrlClient := cc.Client.TenantCluster.K8s.CtrlClient()

	for _, dsName := range []string{awsNodeDsName, KubeProxyDsName} {
		ds := appsv1.DaemonSet{}
		err = ctrlClient.Get(ctx, client.ObjectKey{
			Namespace: dsNamespace,
			Name:      dsName,
		}, &ds)
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "Daemonset %q was not found in namespace %q", dsName, dsNamespace)

			continue
		} else if err != nil {
			return microerror.Mask(err)
		}
		// Check if the daemonset already has the node affinity entry we need.
		aff := ds.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		for _, expression := range aff.NodeSelectorTerms[0].MatchExpressions {
			if expression.Key == label.OperatorVersion &&
				expression.Operator == "NotIn" &&
				expression.Values[0] == project.Version() {

				// Node affinity entry found, nothing to do.
				r.logger.Debugf(ctx, "Daemonset is already restricted")

				continue
			}
		}

		// Node affinity entry missing, add it.
		r.logger.Debugf(ctx, "Daemonset needs to be patched")
		expr := ds.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions
		expr = append(expr, corev1.NodeSelectorRequirement{
			Key:      label.OperatorVersion,
			Operator: "NotIn",
			Values:   []string{project.Version()},
		})

		ds.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions = expr

		err = ctrlClient.Update(ctx, &ds)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "Daemonset patched successfully")
	}

	return nil
}
