package tccp

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/finalizerskeptcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	mdList := &infrastructurev1alpha3.AWSMachineDeploymentList{}
	{
		r.logger.Debugf(ctx, "finding machine deployments for the tenant cluster")

		l := metav1.AddLabelToSelector(
			&metav1.LabelSelector{},
			label.Cluster,
			key.ClusterID(&cr),
		)
		o := metav1.ListOptions{
			LabelSelector: labels.Set(l.MatchLabels).String(),
		}

		err = r.ctrlClient.List(ctx, mdList, &client.ListOptions{Raw: &o})
		if err != nil {
			return microerror.Mask(err)
		}

		if len(mdList.Items) != 0 {
			r.logger.Debugf(ctx, "found %d machine deployments for the tenant cluster", len(mdList.Items))
			r.logger.Debugf(ctx, "not deleting the tenant cluster's control plane cloud formation stack")

			r.logger.Debugf(ctx, "keeping finalizers")
			finalizerskeptcontext.SetKept(ctx)

			return nil
		}

		r.logger.Debugf(ctx, "did not find machine deployments for the tenant cluster")
	}

	{
		r.logger.Debugf(ctx, "disabling the termination protection of the tenant cluster's control plane cloud formation stack")

		i := &cloudformation.UpdateTerminationProtectionInput{
			EnableTerminationProtection: aws.Bool(false),
			StackName:                   aws.String(key.StackNameTCCP(&cr)),
		}

		_, err = cc.Client.TenantCluster.AWS.CloudFormation.UpdateTerminationProtection(i)
		if IsDeleteInProgress(err) {
			r.logger.Debugf(ctx, "the tenant cluster's control plane cloud formation stack is being deleted")
			r.event.Emit(ctx, &cr, "CFDelete", fmt.Sprintf("the tenant cluster's control plane cloud formation stack has stack status %#q", cloudformation.StackStatusDeleteInProgress))

			r.logger.Debugf(ctx, "keeping finalizers")
			finalizerskeptcontext.SetKept(ctx)

			r.logger.Debugf(ctx, "canceling resource")

			return nil
		} else if IsDeleteFailed(err) {
			r.logger.Debugf(ctx, "the tenant cluster's control plane cloud formation stack failed to delete")
			r.event.Emit(ctx, &cr, "CFDeleteFailed", fmt.Sprintf("the tenant cluster's control plane cloud formation stack has stack status %#q", cloudformation.StackStatusDeleteFailed))

			r.logger.Debugf(ctx, "keeping finalizers")
			finalizerskeptcontext.SetKept(ctx)

			r.logger.Debugf(ctx, "canceling resource")

			return nil

		} else if IsNotExists(err) {
			r.logger.Debugf(ctx, "the tenant cluster's control plane cloud formation stack does not exist")
			r.event.Emit(ctx, &cr, "CFDeleted", fmt.Sprintf("the tenant cluster's control plane cloud formation stack has stack status %#q", cloudformation.StackStatusDeleteComplete))
			r.logger.Debugf(ctx, "canceling resource")

			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "disabled the termination protection of the tenant cluster's control plane cloud formation stack")
	}

	{
		r.logger.Debugf(ctx, "requesting the deletion of the tenant cluster's control plane cloud formation stack")

		i := &cloudformation.DeleteStackInput{
			StackName: aws.String(key.StackNameTCCP(&cr)),
		}

		_, err = cc.Client.TenantCluster.AWS.CloudFormation.DeleteStack(i)
		if IsUpdateInProgress(err) {
			r.logger.Debugf(ctx, "the tenant cluster's control plane cloud formation stack is being updated")

			r.logger.Debugf(ctx, "keeping finalizers")
			finalizerskeptcontext.SetKept(ctx)

			r.logger.Debugf(ctx, "canceling resource")

			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "requested the deletion of the tenant cluster's control plane cloud formation stack")
		r.event.Emit(ctx, &cr, "CFDeleteRequested", "requested the deletion of the tenant cluster's control plane cloud formation stack")

		r.logger.Debugf(ctx, "keeping finalizers")
		finalizerskeptcontext.SetKept(ctx)
	}

	return nil
}
