package tccp

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	"github.com/giantswarm/aws-operator/service/controller/v25/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v25/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v25/key"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	stackStateToDelete, err := toStackState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if stackStateToDelete.Name != "" {
		err = r.disableMasterTerminationProtection(ctx, key.MasterInstanceName(cr))
		if err != nil {
			return microerror.Mask(err)
		}

		{
			r.logger.LogCtx(ctx, "level", "debug", "message", "disabling the termination protection of the tenant cluster's control plane cloud formation stack")

			i := &cloudformation.UpdateTerminationProtectionInput{
				EnableTerminationProtection: aws.Bool(false),
				StackName:                   aws.String(key.MainGuestStackName(cr)),
			}

			_, err = cc.Client.TenantCluster.AWS.CloudFormation.UpdateTerminationProtection(i)
			if IsDeleteInProgress(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster's control plane cloud formation stack is being deleted")

				r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")
				finalizerskeptcontext.SetKept(ctx)

				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)

				return nil

			} else if IsNotExists(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster's control plane cloud formation stack does not exist")

				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)

				return nil

			} else if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "disabled the termination protection of the tenant cluster's control plane cloud formation stack")
		}

		{
			r.logger.LogCtx(ctx, "level", "debug", "message", "requesting the deletion of the tenant cluster's control plane cloud formation stack")

			i := &cloudformation.DeleteStackInput{
				StackName: aws.String(key.MainGuestStackName(cr)),
			}

			_, err = cc.Client.TenantCluster.AWS.CloudFormation.DeleteStack(i)
			if err != nil {
				return microerror.Mask(err)
			}

			if r.encrypterBackend == encrypter.VaultBackend {
				err = r.encrypterRoleManager.EnsureDeletedAuthorizedIAMRoles(ctx, cr)
				if err != nil {
					return microerror.Mask(err)
				}
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "requested the deletion of the tenant cluster's control plane cloud formation stack")
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not deleting the tenant cluster main stack")
	}

	return nil
}

func (r *Resource) disableMasterTerminationProtection(ctx context.Context, masterInstanceName string) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "disabling master instance termination protection")

	var reservations []*ec2.Reservation
	{
		i := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("tag:Name"),
					Values: []*string{
						aws.String(masterInstanceName),
					},
				},
			},
		}

		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(o.Reservations) != 1 {
			return microerror.Maskf(executionFailedError, "expected one reservation for master instance, got %d", len(o.Reservations))
		}

		reservations = o.Reservations
	}

	for _, reservation := range reservations {
		if len(reservation.Instances) != 1 {
			return microerror.Maskf(executionFailedError, "expected one master instance, got %d", len(reservation.Instances))
		}

		for _, instance := range reservation.Instances {
			i := &ec2.ModifyInstanceAttributeInput{
				DisableApiTermination: &ec2.AttributeBooleanValue{
					Value: aws.Bool(false),
				},
				InstanceId: aws.String(*instance.InstanceId),
			}

			_, err = cc.Client.TenantCluster.AWS.EC2.ModifyInstanceAttribute(i)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "disabled master instance termination protection")

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	deleteChange, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetDeleteChange(deleteChange)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentStackState, err := toStackState(currentState)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	deleteState := StackState{
		Name: currentStackState.Name,
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster main stack that has to be deleted")

	return deleteState, nil
}
