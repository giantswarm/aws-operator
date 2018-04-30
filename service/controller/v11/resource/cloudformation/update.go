package cloudformation

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/context/updateallowedcontext"

	"github.com/giantswarm/aws-operator/service/controller/v11/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	stackStateToUpdate, err := toStackState(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if stackStateToUpdate.Name != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating the guest cluster main stack")

		if stackStateToUpdate.ShouldUpdate && !stackStateToUpdate.ShouldScale {
			// Fetch the etcd volume information.
			etcdVolume := true
			persistentVolume := false
			volumes, err := r.ebs.ListVolumes(customObject, etcdVolume, persistentVolume)
			if err != nil {
				return microerror.Mask(err)
			}
			if len(volumes) != 1 {
				return microerror.Maskf(executionFailedError, "there must be 1 volume for etcd, got %d", len(volumes))
			}
			vol := volumes[0]

			// First shutdown the instances and wait for it to be stopped. Then detach
			// the etcd volume without forcing.
			force := false
			shutdown := true
			wait := true
			for _, a := range vol.Attachments {
				err := r.ebs.DetachVolume(ctx, vol.VolumeID, a, force, shutdown, wait)
				if err != nil {
					return microerror.Mask(err)
				}
			}
		}

		// Once the etcd volume is cleaned up and the master instance is down we can
		// go ahead to let CloudFormation do its job.
		_, err = r.clients.CloudFormation.UpdateStack(&stackStateToUpdate.UpdateStackInput)
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

func (r *Resource) computeUpdateState(customObject v1alpha1.AWSConfig, stackState StackState) (cloudformation.UpdateStackInput, error) {
	mainTemplate, err := r.getMainGuestTemplateBody(customObject, stackState)
	if err != nil {
		return cloudformation.UpdateStackInput{}, microerror.Mask(err)
	}

	updateStackInput := cloudformation.UpdateStackInput{
		Capabilities: []*string{
			// CAPABILITY_NAMED_IAM is required for updating IAM roles (worker
			// policy).
			aws.String(namedIAMCapability),
		},
		StackName:    aws.String(stackState.Name),
		TemplateBody: aws.String(mainTemplate),
	}

	return updateStackInput, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
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
	if updateallowedcontext.IsUpdateAllowed(ctx) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the guest cluster main stack has to be updated")

		if shouldUpdate(currentStackState, desiredStackState) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack has to be updated")

			updateStackInput, err := r.computeUpdateState(customObject, desiredStackState)
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
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not computing update state of the guest cluster main stack because updates are not allowed")
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

		if shouldScale(currentStackState, desiredStackState) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack has to be scaled")

			desiredStackState.MasterInstanceResourceName = currentStackState.MasterInstanceResourceName

			updateStackInput, err := r.computeUpdateState(customObject, desiredStackState)
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
func shouldScale(currentState, desiredState StackState) bool {
	if currentState.MasterImageID != desiredState.MasterImageID {
		return false
	}
	if currentState.MasterInstanceType != desiredState.MasterInstanceType {
		return false
	}
	if currentState.MasterCloudConfigVersion != desiredState.MasterCloudConfigVersion {
		return false
	}
	if currentState.WorkerImageID != desiredState.WorkerImageID {
		return false
	}
	if currentState.WorkerInstanceType != desiredState.WorkerInstanceType {
		return false
	}
	if currentState.WorkerCloudConfigVersion != desiredState.WorkerCloudConfigVersion {
		return false
	}
	if currentState.VersionBundleVersion != desiredState.VersionBundleVersion {
		return false
	}

	if currentState.WorkerCount != desiredState.WorkerCount {
		return true
	}

	return false
}

// shouldUpdate determines whether the reconciled guest cluster should be
// updated. A guest cluster is only allowed to update in the following cases.
//
//     The instance type of master nodes changes (indicates updates).
//     The instance type of worker nodes changes (indicates updates).
//     The version bundle version changes (indicates updates).
//
func shouldUpdate(currentState, desiredState StackState) bool {
	if currentState.MasterInstanceType != desiredState.MasterInstanceType {
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
