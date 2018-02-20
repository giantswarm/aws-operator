package cloudformation

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/context/updateallowedcontext"

	"github.com/giantswarm/aws-operator/service/awsconfig/v5/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	updateStackInput, err := toUpdateStackInput(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	stackName := updateStackInput.StackName
	if stackName != nil && *stackName != "" {
		_, err := r.clients.CloudFormation.UpdateStack(&updateStackInput)
		if err != nil {
			return microerror.Maskf(err, "updating AWS cloudformation stack")
		}

		r.logger.LogCtx(ctx, "debug", "updating AWS cloudformation stack: updated")
	} else {
		r.logger.LogCtx(ctx, "debug", "updating AWS cloudformation stack: no need to update")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) computeUpdateState(customObject v1alpha1.AWSConfig, desiredState StackState) (cloudformation.UpdateStackInput, error) {
	mainTemplate, err := r.getMainGuestTemplateBody(customObject)
	if err != nil {
		return cloudformation.UpdateStackInput{}, microerror.Mask(err)
	}

	updateState := cloudformation.UpdateStackInput{
		Capabilities: []*string{
			// CAPABILITY_NAMED_IAM is required for updating IAM roles (worker
			// policy).
			aws.String(namedIAMCapability),
		},
		StackName:    aws.String(desiredState.Name),
		TemplateBody: aws.String(mainTemplate),
	}

	return updateState, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}
	desiredStackState, err := toStackState(desiredState)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}
	currentStackState, err := toStackState(currentState)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	if shouldScale(currentStackState, desiredStackState) {
		r.logger.LogCtx(ctx, "debug", "main stack has to be scaled")

		updateState, err := r.computeUpdateState(customObject, desiredStackState)
		if err != nil {
			return cloudformation.CreateStackInput{}, microerror.Mask(err)
		}

		return updateState, nil
	} else {
		r.logger.LogCtx(ctx, "debug", "main stack has not to be scaled")
	}

	// We enable/disable updates in order to enable them our test installations
	// but disable them in production installations. That is useful until we have
	// full confidence in updating guest clusters.
	if updateallowedcontext.IsUpdateAllowed(ctx) {
		r.logger.LogCtx(ctx, "debug", "finding out if the main stack has to be updated")

		if shouldUpdate(currentStackState, desiredStackState) {
			r.logger.LogCtx(ctx, "debug", "main stack has to be updated")

			updateState, err := r.computeUpdateState(customObject, desiredStackState)
			if err != nil {
				return cloudformation.CreateStackInput{}, microerror.Mask(err)
			}

			return updateState, nil
		} else {
			r.logger.LogCtx(ctx, "debug", "main stack has not to be updated")
		}
	} else {
		r.logger.LogCtx(ctx, "debug", "not computing update state because main stack are not allowed to be updated")
	}

	return cloudformation.UpdateStackInput{}, nil
}

// shouldScale determines whether the reconciled guest cluster should be scaled.
// A guest cluster is only allowed to scale in case nothing but the worker count
// changes. In case anything else changes as well, scaling is not allowed, since
// any other changes should be covered by general updates, which is a separate
// step.
func shouldScale(currentState, desiredState StackState) bool {
	if currentState.Name != desiredState.Name {
		return false
	}
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
//     The update does not indicate scaling.
//     The version bundle version changes
//     The instance type of master nodes changes.
//     The instance type of worker nodes changes.
//
func shouldUpdate(currentState, desiredState StackState) bool {
	if currentState.WorkerCount != desiredState.WorkerCount {
		return false
	}

	if currentState.VersionBundleVersion != desiredState.VersionBundleVersion {
		return true
	}
	if currentState.MasterInstanceType != desiredState.MasterInstanceType {
		return true
	}
	if currentState.WorkerInstanceType != desiredState.WorkerInstanceType {
		return true
	}

	return false
}
