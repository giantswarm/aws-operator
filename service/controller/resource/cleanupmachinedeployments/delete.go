package cleanupmachinedeployments

import (
	"context"
	"fmt"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/finalizerskeptcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var mdList *infrastructurev1alpha3.AWSMachineDeploymentList
	{
		r.logger.Debugf(ctx, "finding AWSMachineDeployments for tenant cluster")

		o := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", label.Cluster, key.ClusterID(&cr)),
		}

		mdList, err = r.g8sClient.InfrastructureV1alpha3().AWSMachineDeployments(metav1.NamespaceAll).List(ctx, o)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found %d AWSMachineDeployments for tenant cluster", len(mdList.Items))
	}

	// We do not want to delete the AWSCluster CR as long as there are any
	// AWSMachineDeployment CRs. This is because there cannot be any Node Pool
	// without a Cluster.
	if len(mdList.Items) != 0 {
		r.logger.Debugf(ctx, "keeping finalizers")
		finalizerskeptcontext.SetKept(ctx)
	}

	for _, md := range mdList.Items {
		r.logger.Debugf(ctx, "deleting aws machine deployment %#q for tenant cluster %#q", md.Namespace+"/"+md.Name, key.ClusterID(&cr))

		err = r.g8sClient.InfrastructureV1alpha3().AWSMachineDeployments(md.Namespace).Delete(ctx, md.Name, metav1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "deleted aws machine deployment %#q for tenant cluster %#q", md.Namespace+"/"+md.Name, key.ClusterID(&cr))
		r.event.Emit(ctx, &cr, "MachineDeploymentDeleted", fmt.Sprintf("deleted aws machine deployment %#q for tenant cluster %#q", md.Namespace+"/"+md.Name, key.ClusterID(&cr)))
	}

	return nil
}
