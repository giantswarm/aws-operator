package cloudformation

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	"github.com/giantswarm/aws-operator/service/controller/v15/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v15/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v15/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	stackInput, err := toCreateStackInput(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if stackInput.StackName != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating the guest cluster main stack")

		sc, err := controllercontext.FromContext(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		if r.encrypterBackend == encrypter.VaultBackend {
			customObject, err := key.ToCustomObject(obj)
			if err != nil {
				return microerror.Mask(err)
			}

			err = r.addRoleAccess(sc, customObject)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		customObject, err := key.ToCustomObject(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		stackInput.Parameters = []*cloudformation.Parameter{
			{
				ParameterKey:   aws.String(versionBundleVersionParameterKey),
				ParameterValue: aws.String(key.VersionBundleVersion(customObject)),
			},
		}

		_, err = sc.AWSClient.CloudFormation.CreateStack(&stackInput)
		if IsAlreadyExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		err = sc.AWSClient.CloudFormation.WaitUntilStackCreateCompleteWithContext(ctx, &cloudformation.DescribeStacksInput{
			StackName: stackInput.StackName,
		})
		if ctx.Err() == context.DeadlineExceeded {
			// We waited longer than we wanted to get a reasonable result and be sure
			// the stack got properly created. We skip here and try again on the next
			// resync.
			r.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster main stack creation is not complete")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")

			return nil
		} else if IsResourceNotReady(err) {
			// There might be cases in which AWS is not fast enough to create the
			// resources we want to watch. We skip here and try again on the next
			// resync.
			r.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster main stack creation is not complete")
			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")

			return nil
		} else if ctx.Err() != nil {
			return microerror.Mask(ctx.Err())
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "created the guest cluster main stack")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not creating the guest cluster main stack")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
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

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the guest cluster main stack has to be created")

	createState := cloudformation.CreateStackInput{}

	if currentStackState.Name == "" || desiredStackState.Name != currentStackState.Name {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack has to be created")

		if err := r.validateCluster(customObject); err != nil {
			return cloudformation.CreateStackInput{}, microerror.Mask(err)
		}

		// We need to create the required peering resources in the host account before
		// getting the guest main stack template body, it requires id values from host
		// resources.
		err = r.createHostPreStack(ctx, customObject)
		if err != nil {
			return cloudformation.CreateStackInput{}, microerror.Mask(err)
		}

		var mainTemplate string
		mainTemplate, err := r.getMainGuestTemplateBody(ctx, customObject, desiredStackState)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		createState.StackName = aws.String(desiredStackState.Name)
		createState.TemplateBody = aws.String(mainTemplate)
		createState.TimeoutInMinutes = aws.Int64(defaultCreationTimeout)
		// CAPABILITY_NAMED_IAM is required for creating IAM roles (worker policy)
		createState.Capabilities = []*string{
			aws.String(namedIAMCapability),
		}

		createState.SetTags(r.getCloudFormationTags(customObject))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack does not have to be created")

		// Create host post-main stack once the guest main stack got created. This
		// here usually happens on the second or third attempt dependening on the
		// resnyc period. It includes the peering routes, which need resources from
		// the guest stack to be in place before it can be created.
		err = r.createHostPostStack(ctx, customObject, currentStackState)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return createState, nil
}

func (r *Resource) createHostPreStack(ctx context.Context, customObject v1alpha1.AWSConfig) error {
	stackName := key.MainHostPreStackName(customObject)
	mainTemplate, err := r.getMainHostPreTemplateBody(ctx, customObject)
	if err != nil {
		return microerror.Mask(err)
	}
	createStack := &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(mainTemplate),
		// CAPABILITY_NAMED_IAM is required for creating IAM roles (worker policy)
		Capabilities: []*string{
			aws.String(namedIAMCapability),
		},
	}
	createStack.SetTags(r.getCloudFormationTags(customObject))

	r.logger.LogCtx(ctx, "level", "debug", "message", "creating the host cluster pre cloud formation stack")

	_, err = r.hostClients.CloudFormation.CreateStack(createStack)
	if IsAlreadyExists(err) {
		// TODO this here indicates we should have dedicated resources for the pre,
		// main and post stacks. The workflow would be more straight forward and
		// easy to manage. Right now we hack around this and add conditionals to
		// make it work somehow while bypassing the framework primitives.
		r.logger.LogCtx(ctx, "level", "debug", "message", "the host cluster pre cloud formation stack is already created")

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	err = r.hostClients.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "created the host cluster pre cloud formation stack")

	return nil
}

func (r *Resource) createHostPostStack(ctx context.Context, customObject v1alpha1.AWSConfig, guestMainStackState StackState) error {
	stackName := key.MainHostPostStackName(customObject)
	mainTemplate, err := r.getMainHostPostTemplateBody(ctx, customObject, guestMainStackState)
	if err != nil {
		return microerror.Mask(err)
	}
	createStack := &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(mainTemplate),
	}
	createStack.SetTags(r.getCloudFormationTags(customObject))

	r.logger.LogCtx(ctx, "level", "debug", "message", "creating the host cluster post cloud formation stack")

	_, err = r.hostClients.CloudFormation.CreateStack(createStack)
	if IsAlreadyExists(err) {
		// TODO this here indicates we should have dedicated resources for the pre,
		// main and post stacks. The workflow would be more straight forward and
		// easy to manage. Right now we hack around this and add conditionals to
		// make it work somehow while bypassing the framework primitives.
		r.logger.LogCtx(ctx, "level", "debug", "message", "the host cluster post cloud formation stack is already created")

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "created the host cluster post cloud formation stack")

	return nil
}
