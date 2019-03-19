package tccp

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	"github.com/giantswarm/aws-operator/service/controller/v25/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v25/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v25/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	stackStateToCreate, err := toStackState(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if stackStateToCreate.Name != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating the tenant cluster main stack")

		if r.encrypterBackend == encrypter.VaultBackend {
			err = r.encrypterRoleManager.EnsureCreatedAuthorizedIAMRoles(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		{
			i := &cloudformation.CreateStackInput{
				// CAPABILITY_NAMED_IAM is required for creating worker policy IAM roles.
				Capabilities: []*string{
					aws.String(namedIAMCapability),
				},
				EnableTerminationProtection: aws.Bool(key.EnableTerminationProtection),
				Parameters: []*cloudformation.Parameter{
					{
						ParameterKey:   aws.String(versionBundleVersionParameterKey),
						ParameterValue: aws.String(key.VersionBundleVersion(cr)),
					},
				},
				StackName:        aws.String(key.MainGuestStackName(cr)),
				Tags:             r.getCloudFormationTags(cr),
				TemplateBody:     aws.String(stackStateToCreate.Template),
				TimeoutInMinutes: aws.Int64(20),
			}

			_, err = cc.Client.TenantCluster.AWS.CloudFormation.CreateStack(i)
			if IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		err = cc.Client.TenantCluster.AWS.CloudFormation.WaitUntilStackCreateCompleteWithContext(ctx, &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.MainGuestStackName(cr)),
		})
		if ctx.Err() == context.DeadlineExceeded {
			// We waited longer than we wanted to get a reasonable result and be sure
			// the stack got properly created. We skip here and try again on the next
			// resync.
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster main stack creation is not complete")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")

			return nil
		} else if IsResourceNotReady(err) {
			// There might be cases in which AWS is not fast enough to create the
			// resources we want to watch. We skip here and try again on the next
			// resync.
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster main stack creation is not complete")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")

			return nil
		} else if ctx.Err() != nil {
			return microerror.Mask(ctx.Err())
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "created the tenant cluster main stack")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not creating the tenant cluster main stack")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}
	currentStackState, err := toStackState(currentState)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}
	desiredStackState, err := toStackState(desiredState)
	if err != nil {
		return cloudformation.CreateStackInput{}, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the tenant cluster main stack has to be created")

	var createState StackState
	if currentStackState.Name == "" || desiredStackState.Name != currentStackState.Name {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster main stack has to be created")

		templateBody, err := r.newTemplateBody(ctx, cr, desiredStackState)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		createState = StackState{
			Name:     desiredStackState.Name,
			Template: templateBody,
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster main stack does not have to be created")
	}

	return createState, nil
}
