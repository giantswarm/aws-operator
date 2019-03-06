package cloudformation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"

	"github.com/giantswarm/aws-operator/service/controller/v24/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v24/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v24/key"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	stackStateToDelete, err := toStackState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if stackStateToDelete.Name != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting the guest cluster main stack")

		cc, err := controllercontext.FromContext(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		input := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("tag:Name"),
					Values: []*string{
						aws.String(key.MasterInstanceName(customObject)),
					},
				},
			},
		}
		result, err := cc.AWSClient.EC2.DescribeInstances(input)
		fmt.Println(result)
		if IsNotExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			for _, reservation := range result.Reservations {
				for _, instance := range reservation.Instances {
					r.logger.LogCtx(ctx, "level", "debug", "message", "disabling master instance termination protection")
					input := &ec2.ModifyInstanceAttributeInput{
						DisableApiTermination: &ec2.AttributeBooleanValue{
							Value: aws.Bool(false),
						},
						InstanceId: aws.String(*instance.InstanceId),
					}

					_, err = cc.AWSClient.EC2.ModifyInstanceAttribute(input)
					if IsNotExists(err) {
						// fall through
					} else if err != nil {
						return microerror.Mask(err)
					}

					r.logger.LogCtx(ctx, "level", "debug", "message", "disabled master instance termination protection")
				}
			}
		}

		stackName := aws.String(key.MainGuestStackName(customObject))

		updateTerminationProtection := &cloudformation.UpdateTerminationProtectionInput{
			EnableTerminationProtection: aws.Bool(false),
			StackName:                   stackName,
		}
		_, err = cc.AWSClient.CloudFormation.UpdateTerminationProtection(updateTerminationProtection)
		if IsDeleteInProgress(err) {
			// fall through
		} else if IsNotExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			i := &cloudformation.DeleteStackInput{
				StackName: stackName,
			}
			_, err = cc.AWSClient.CloudFormation.DeleteStack(i)
			if err != nil {
				return microerror.Mask(err)
			}

			if r.encrypterBackend == encrypter.VaultBackend {
				err = r.encrypterRoleManager.EnsureDeletedAuthorizedIAMRoles(ctx, customObject)
				if err != nil {
					return microerror.Mask(err)
				}
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "deleted the guest cluster main stack")
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not deleting the guest cluster main stack")
	}

	if stackStateToDelete.Name != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting the host cluster post stack")

		stackName := aws.String(key.MainHostPostStackName(customObject))

		updateTerminationProtection := &cloudformation.UpdateTerminationProtectionInput{
			EnableTerminationProtection: aws.Bool(false),
			StackName:                   stackName,
		}
		_, err = r.hostClients.CloudFormation.UpdateTerminationProtection(updateTerminationProtection)
		if IsDeleteInProgress(err) {
			// fall through
		} else if IsNotExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			i := &cloudformation.DeleteStackInput{
				StackName: stackName,
			}
			_, err = r.hostClients.CloudFormation.DeleteStack(i)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", "deleted the host cluster post stack")
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not deleting the host cluster post stack")
	}

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

	r.logger.LogCtx(ctx, "level", "debug", "message", "found the guest cluster main stack that has to be deleted")

	return deleteState, nil
}
