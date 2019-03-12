package tccp

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"

	"github.com/giantswarm/aws-operator/service/controller/v25/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v25/ebs"
	"github.com/giantswarm/aws-operator/service/controller/v25/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	stackStateToUpdate, err := toStackState(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	var ebsService ebs.Interface
	{
		c := ebs.Config{
			Client: cc.Client.TenantCluster.AWS.EC2,
			Logger: r.logger,
		}

		ebsService, err = ebs.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	if stackStateToUpdate.Name != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating the guest cluster main stack")

		if stackStateToUpdate.ShouldUpdate && !stackStateToUpdate.ShouldScale {
			{
				// Fetch the etcd volume information.
				filterFuncs := []func(t *ec2.Tag) bool{
					ebs.NewDockerVolumeFilter(cr),
					ebs.NewEtcdVolumeFilter(cr),
				}
				volumes, err := ebsService.ListVolumes(cr, filterFuncs...)
				if err != nil {
					return microerror.Mask(err)
				}

				// First shutdown the instances and wait for it to be stopped. Then detach
				// the etcd and docker volume without forcing.
				force := false
				shutdown := true
				wait := true

				for _, v := range volumes {
					for _, a := range v.Attachments {
						err := ebsService.DetachVolume(ctx, v.VolumeID, a, force, shutdown, wait)
						if err != nil {
							return microerror.Mask(err)
						}
					}
				}
			}

			{
				err := r.terminateOldMasterInstance(ctx, obj)
				if err != nil {
					return microerror.Mask(err)
				}
			}
		}

		// Once the etcd volume is cleaned up and the master instance is down we can
		// go ahead to let CloudFormation do its job.
		_, err = cc.Client.TenantCluster.AWS.CloudFormation.UpdateStack(&stackStateToUpdate.UpdateStackInput)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated the guest cluster main stack")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not updating the guest cluster main stack")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) computeUpdateState(ctx context.Context, cr v1alpha1.AWSConfig, stackState StackState) (cloudformation.UpdateStackInput, error) {
	mainTemplate, err := r.getMainGuestTemplateBody(ctx, cr, stackState)
	if err != nil {
		return cloudformation.UpdateStackInput{}, microerror.Mask(err)
	}

	updateStackInput := cloudformation.UpdateStackInput{
		Capabilities: []*string{
			// CAPABILITY_NAMED_IAM is required for updating IAM roles (worker
			// policy).
			aws.String(namedIAMCapability),
		},
		Parameters: []*cloudformation.Parameter{
			{
				ParameterKey:   aws.String(versionBundleVersionParameterKey),
				ParameterValue: aws.String(key.VersionBundleVersion(cr)),
			},
		},
		StackName:    aws.String(stackState.Name),
		TemplateBody: aws.String(mainTemplate),
	}

	return updateStackInput, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}
	currentStackState, err := toStackState(currentState)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}
	desiredStackState, err := toStackState(desiredState)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	// We enable/disable updates in order to enable them our test installations
	// but disable them in production installations. That is useful until we have
	// full confidence in updating guest clusters. Note that updates also manage
	// scaling at the same time to be more efficient.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the guest cluster main stack has to be updated")

		if shouldUpdate(currentStackState, desiredStackState) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack has to be updated")

			updateStackInput, err := r.computeUpdateState(ctx, cr, desiredStackState)
			if err != nil {
				return StackState{}, microerror.Mask(err)
			}

			updateState := StackState{
				Name:             desiredStackState.Name,
				ShouldScale:      false,
				ShouldUpdate:     true,
				UpdateStackInput: updateStackInput,
			}

			return updateState, nil
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack does not have to be updated")
		}
	}

	// We manage scaling separately because the impact and implications of scaling
	// is different compared to updates. We can just process scaling any time. We
	// cannot just process updates at any time and thus have to separate the
	// management of both primitives. Note that updates also manage scaling at the
	// same time for more efficiency. Note that we have to preserve the master
	// instance resource name when scaling worker nodes to prevent updating the
	// master node. This is why we set the desired state of the master instance
	// resource name to the current state below.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the guest cluster main stack has to be scaled")

		shouldScale, err := r.shouldScale(ctx, cr, currentStackState, desiredStackState)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		if shouldScale {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack has to be scaled")

			desiredStackState.MasterInstanceResourceName = currentStackState.MasterInstanceResourceName
			desiredStackState.DockerVolumeResourceName = currentStackState.DockerVolumeResourceName

			updateStackInput, err := r.computeUpdateState(ctx, cr, desiredStackState)
			if err != nil {
				return StackState{}, microerror.Mask(err)
			}

			updateState := StackState{
				Name:             desiredStackState.Name,
				ShouldScale:      true,
				ShouldUpdate:     false,
				UpdateStackInput: updateStackInput,
			}

			return updateState, nil
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack does not have to be scaled")
		}
	}

	return StackState{}, nil
}

// shouldScale determines whether the reconciled guest cluster should be scaled.
// A guest cluster is only allowed to scale in case nothing but the worker count
// changes. In case anything else changes as well, scaling is not allowed, since
// any other changes should be covered by general updates, which is a separate
// step.
func (r *Resource) shouldScale(ctx context.Context, cr v1alpha1.AWSConfig, currentState, desiredState StackState) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if currentState.MasterImageID != desiredState.MasterImageID {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not scaling due to master image id")
		return false, nil
	}
	if currentState.MasterInstanceType != desiredState.MasterInstanceType {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not scaling due to master instance type")
		return false, nil
	}
	if currentState.MasterCloudConfigVersion != desiredState.MasterCloudConfigVersion {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not scaling due to master cloudconfig version")
		return false, nil
	}
	if currentState.WorkerDockerVolumeSizeGB != desiredState.WorkerDockerVolumeSizeGB {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not scaling due to worker docker volume size")
		return false, nil
	}
	if currentState.WorkerImageID != desiredState.WorkerImageID {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not scaling due to worker image id")
		return false, nil
	}
	if currentState.WorkerInstanceType != desiredState.WorkerInstanceType {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not scaling due to worker instance type")
		return false, nil
	}
	if currentState.WorkerCloudConfigVersion != desiredState.WorkerCloudConfigVersion {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not scaling due to worker cloudconfig version")
		return false, nil
	}
	if !cc.Status.TenantCluster.TCCP.ASG.IsEmpty() && cc.Status.TenantCluster.TCCP.ASG.MaxSize != key.ScalingMax(cr) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "scaling due to scaling.max")
		return true, nil
	}
	if !cc.Status.TenantCluster.TCCP.ASG.IsEmpty() && cc.Status.TenantCluster.TCCP.ASG.MinSize != key.ScalingMin(cr) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "scaling due to scaling.min")
		return true, nil
	}
	if currentState.VersionBundleVersion != desiredState.VersionBundleVersion {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not scaling due to version bundle version")
		return false, nil
	}

	return false, nil
}

// shouldUpdate determines whether the reconciled guest cluster should be
// updated. A guest cluster is only allowed to update in the following cases.
//
//     The instance type of master nodes changes (indicates updates).
//     The instance type of worker nodes changes (indicates updates).
//     The size of a docker volume for worker nodes changes.
//     The version bundle version changes (indicates updates).
//
func shouldUpdate(currentState, desiredState StackState) bool {
	if currentState.MasterInstanceType != desiredState.MasterInstanceType {
		return true
	}
	if currentState.WorkerDockerVolumeSizeGB != desiredState.WorkerDockerVolumeSizeGB {
		return true
	}
	if currentState.WorkerInstanceType != desiredState.WorkerInstanceType {
		return true
	}
	if currentState.VersionBundleVersion != desiredState.VersionBundleVersion {
		return true
	}

	return false
}

// Terminates the master instance of the cluster.
//
// To detect the old master instance we find the instance by its name and
// the instance state "stopped". Within the upgrade process the master first
// gets stopped and its volumes get detached. This function makes sure that
// the stopped instance is also terminated.
func (r *Resource) terminateOldMasterInstance(ctx context.Context, obj interface{}) error {
	var result *ec2.DescribeInstancesOutput

	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	instanceName := key.MasterInstanceName(cr)
	instanceState := "stopped"

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding master instance with name %#q in state %#q", instanceName, instanceState))

		i := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("tag:Name"),
					Values: []*string{
						aws.String(instanceName),
					},
				},
				{
					Name: aws.String("instance-state-name"),
					Values: []*string{
						aws.String(instanceState),
					},
				},
				{
					Name: aws.String("tag:giantswarm.io/cluster"),
					Values: []*string{
						aws.String(key.ClusterID(cr)),
					},
				},
			},
		}

		result, err = cc.Client.TenantCluster.AWS.EC2.DescribeInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(result.Reservations) > 1 {
			return microerror.Maskf(executionFailedError, "expected at most one master instance with name %#q in state %#q but got %d", instanceName, instanceState, len(result.Reservations))
		}

		if len(result.Reservations) < 1 {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find master instance with name %#q in state %#q", instanceName, instanceState))
			return nil
		}

		if len(result.Reservations[0].Instances) > 1 {
			return microerror.Maskf(executionFailedError, "expected at most one master instance with name %#q in state %#q but got %d", instanceName, instanceState, len(result.Reservations))
		}

		if len(result.Reservations[0].Instances) < 1 {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find master instance with state `stopped` and name %#q", key.MasterInstanceName(cr)))
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found master instance with name %#q in state %#q", instanceName, instanceState))
	}

	{
		instanceID := *result.Reservations[0].Instances[0].InstanceId

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("terminating master instance with ID %#q", instanceID))

		i := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{
				aws.String(instanceID),
			},
		}
		_, err := cc.Client.TenantCluster.AWS.EC2.TerminateInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("terminated master instance with ID %#q", instanceID))
	}

	return nil
}
