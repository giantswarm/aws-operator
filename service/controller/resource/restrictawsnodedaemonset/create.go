package restrictawsnodedaemonset

import (
	"context"

	"github.com/giantswarm/microerror"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/v14/pkg/label"
	"github.com/giantswarm/aws-operator/v14/pkg/project"
	"github.com/giantswarm/aws-operator/v14/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v14/service/controller/key"
)

const (
	dsNamespace     = "kube-system"
	awsNodeDsName   = "aws-node"
	kubeProxyDsName = "kube-proxy"
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

	// Only run this if the Cluster CR has the cilium pod annotation
	cluster := apiv1beta1.Cluster{}
	err = r.ctrlClient.Get(ctx, client.ObjectKey{Namespace: cr.Namespace, Name: key.ClusterID(&cr)}, &cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	daemonsetNames := []string{awsNodeDsName}
	if key.IsCiliumEniModeEnabled(cluster) {
		daemonsetNames = append(daemonsetNames, kubeProxyDsName)
	}

	for _, dsName := range daemonsetNames {
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
		requirements := make([]corev1.NodeSelectorRequirement, 0)

		// Check if the daemonset already has the node affinity entry we need.
		if ds.Spec.Template.Spec.Affinity != nil &&
			ds.Spec.Template.Spec.Affinity.NodeAffinity != nil &&
			ds.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil &&
			len(ds.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) > 0 {

			requirements = ds.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions
		}

		newRequirements, changed := ensureAndFilterNodeSelectorRequirements(requirements)
		if !changed {
			r.logger.Debugf(ctx, "Daemonset %q is already filtered to only run on old nodes", dsName)
			return nil
		}

		// Node affinity entry missing, add it.
		r.logger.Debugf(ctx, "Daemonset %q needs to be patched", dsName)

		if ds.Spec.Template.Spec.Affinity == nil {
			ds.Spec.Template.Spec.Affinity = &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: newRequirements,
							},
						},
					},
				},
			}
		} else {
			ds.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions = newRequirements
		}

		err = ctrlClient.Update(ctx, &ds)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "Daemonset %q patched successfully", dsName)
	}

	return nil
}

// ensureAndFilterNodeSelectorRequirements looks for a `requirement` having key == "aws-operator.giantswarm.io/version" and operator == "NotIn"
// and ensures it has the corect value (the project version of the running aws operator). It leaves all other requirements untouched.
func ensureAndFilterNodeSelectorRequirements(requirements []corev1.NodeSelectorRequirement) ([]corev1.NodeSelectorRequirement, bool) {
	ret := make([]corev1.NodeSelectorRequirement, 0)
	changed := false
	found := false

	for _, requirement := range requirements {
		// Check if Key or Operator are not the ones we're looking for.
		if requirement.Key != label.OperatorVersion || requirement.Operator != "NotIn" {
			// This requirement is acting on a different label than the one we care, keep it as-is.
			ret = append(ret, requirement)
		} else {
			// This requirement is using the same key and operator as the one we expected to set, so let's check if the value matches.
			if len(requirement.Values) != 1 || requirement.Values[0] != project.Version() {
				// Requirement is not valid, let's remove it.
				changed = true
			} else {
				// Requirement as we wanted is already there.
				found = true
				ret = append(ret, requirement)
			}
		}
	}

	if !found {
		ret = append(ret, corev1.NodeSelectorRequirement{
			Key:      label.OperatorVersion,
			Operator: "NotIn",
			Values:   []string{project.Version()},
		})

		changed = true
	}

	return ret, changed
}
