package tccp

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v24/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v24/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v24/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	stackInput, err := toCreateStackInput(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if stackInput.StackName != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating the guest cluster main stack")

		cc, err := controllercontext.FromContext(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		if r.encrypterBackend == encrypter.VaultBackend {
			customObject, err := key.ToCustomObject(obj)
			if err != nil {
				return microerror.Mask(err)
			}

			err = r.encrypterRoleManager.EnsureCreatedAuthorizedIAMRoles(ctx, customObject)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		_, err = cc.AWSClient.CloudFormation.CreateStack(&stackInput)
		if IsAlreadyExists(err) {
			// fall through
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

	var createState cloudformation.CreateStackInput
	if currentStackState.Name == "" || desiredStackState.Name != currentStackState.Name {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack has to be created")

		var mainTemplate string
		mainTemplate, err := r.getMainGuestTemplateBody(ctx, customObject, desiredStackState)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		createState = cloudformation.CreateStackInput{
			// CAPABILITY_NAMED_IAM is required for creating worker policy IAM roles.
			Capabilities: []*string{
				aws.String(namedIAMCapability),
			},
			EnableTerminationProtection: aws.Bool(key.EnableTerminationProtection),
			Parameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String(versionBundleVersionParameterKey),
					ParameterValue: aws.String(key.VersionBundleVersion(customObject)),
				},
			},
			StackName:        aws.String(desiredStackState.Name),
			Tags:             r.getCloudFormationTags(customObject),
			TemplateBody:     aws.String(mainTemplate),
			TimeoutInMinutes: aws.Int64(defaultCreationTimeout),
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack does not have to be created")
	}

	return createState, nil
}
