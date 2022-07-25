package cleanupmachinedeployments

import (
	"context"
	"fmt"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/finalizerskeptcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/v13/pkg/label"
	"github.com/giantswarm/aws-operator/v13/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	mdList := &infrastructurev1alpha3.AWSMachineDeploymentList{}
	{
		r.logger.Debugf(ctx, "finding AWSMachineDeployments for tenant cluster")

		o := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", label.Cluster, key.ClusterID(&cr)),
		}

		err = r.ctrlClient.List(ctx, mdList, &client.ListOptions{Raw: &o})
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

	for i, md := range mdList.Items {
		r.logger.Debugf(ctx, "deleting aws machine deployment %#q for tenant cluster %#q", md.Namespace+"/"+md.Name, key.ClusterID(&cr))

		err = r.ctrlClient.Delete(ctx, &mdList.Items[i], &client.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "deleted aws machine deployment %#q for tenant cluster %#q", md.Namespace+"/"+md.Name, key.ClusterID(&cr))
		r.event.Emit(ctx, &cr, "MachineDeploymentDeleted", fmt.Sprintf("deleted aws machine deployment %#q for tenant cluster %#q", md.Namespace+"/"+md.Name, key.ClusterID(&cr)))
	}

	return nil
}
