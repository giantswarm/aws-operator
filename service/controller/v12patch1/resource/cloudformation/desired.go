package cloudformation

import (
	"context"
	"strconv"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v12patch1/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v12patch1/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var mainStack StackState
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing desired state for the guest cluster main stack")

		imageID, err := key.ImageID(customObject)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		// FIXME: the instance type should not depend on the number of workers.
		// issue: https://github.com/giantswarm/awstpr/issues/47
		var workerInstanceType string
		if key.WorkerCount(customObject) > 0 {
			workerInstanceType = key.WorkerInstanceType(customObject)
		}

		var masterInstanceType string
		if len(customObject.Spec.AWS.Masters) > 0 {
			masterInstanceType = key.MasterInstanceType(customObject)
		}

		mainStack = StackState{
			Name: key.MainGuestStackName(customObject),

			MasterImageID:              imageID,
			MasterInstanceResourceName: key.MasterInstanceResourceName(customObject),
			MasterInstanceType:         masterInstanceType,
			MasterCloudConfigVersion:   cloudconfig.CloudConfigVersion,
			MasterInstanceMonitoring:   r.monitoring,

			WorkerCount:              strconv.Itoa(key.WorkerCount(customObject)),
			WorkerImageID:            imageID,
			WorkerInstanceMonitoring: r.monitoring,
			WorkerInstanceType:       workerInstanceType,
			WorkerCloudConfigVersion: cloudconfig.CloudConfigVersion,

			VersionBundleVersion: key.VersionBundleVersion(customObject),
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed desired state for the guest cluster main stack")
	}

	return mainStack, nil
}
