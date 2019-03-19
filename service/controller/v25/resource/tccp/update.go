package tccp

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
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
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating the tenant cluster main stack")

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

			err = r.terminateMasterInstance(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		// Once the etcd volume is cleaned up and the master instance is down we can
		// go ahead to let CloudFormation do its job.
		{
			i := &cloudformation.UpdateStackInput{
				Capabilities: []*string{
					// CAPABILITY_NAMED_IAM is required for updating worker policy IAM
					// roles.
					aws.String(namedIAMCapability),
				},
				Parameters: []*cloudformation.Parameter{
					{
						ParameterKey:   aws.String(versionBundleVersionParameterKey),
						ParameterValue: aws.String(key.VersionBundleVersion(cr)),
					},
				},
				StackName:    aws.String(stackStateToUpdate.Name),
				TemplateBody: aws.String(stackStateToUpdate.Template),
			}

			_, err = cc.Client.TenantCluster.AWS.CloudFormation.UpdateStack(i)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated the tenant cluster main stack")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not updating the tenant cluster main stack")
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
	// full confidence in updating tenant clusters. Note that updates also manage
	// scaling at the same time to be more efficient.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the tenant cluster main stack has to be updated")

		shouldUpdate, err := r.detection.ShouldUpdate(ctx, cr)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		if shouldUpdate {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster main stack has to be updated")

			templateBody, err := r.newTemplateBody(ctx, cr, desiredStackState)
			if err != nil {
				return StackState{}, microerror.Mask(err)
			}

			updateState := StackState{
				Name:     desiredStackState.Name,
				Template: templateBody,

				ShouldScale:  false,
				ShouldUpdate: true,
			}

			return updateState, nil
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster main stack does not have to be updated")
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
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the tenant cluster main stack has to be scaled")

		shouldScale, err := r.detection.ShouldScale(ctx, cr)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		if shouldScale {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster main stack has to be scaled")

			desiredStackState.MasterInstanceResourceName = currentStackState.MasterInstanceResourceName
			desiredStackState.DockerVolumeResourceName = currentStackState.DockerVolumeResourceName

			templateBody, err := r.newTemplateBody(ctx, cr, desiredStackState)
			if err != nil {
				return StackState{}, microerror.Mask(err)
			}

			updateState := StackState{
				Name:     desiredStackState.Name,
				Template: templateBody,

				ShouldScale:  true,
				ShouldUpdate: false,
			}

			return updateState, nil
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster main stack does not have to be scaled")
		}
	}

	return StackState{}, nil
}
