package cpf

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v24/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v24/key"
	"github.com/giantswarm/aws-operator/service/controller/v24/templates"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's control plane finalizer cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.MainHostPreStackName(cr)),
		}

		_, err = r.hostClients.CloudFormation.DescribeStacks(i)
		if IsNotExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's control plane finalizer cloud formation stack already exists")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's control plane finalizer cloud formation stack")
	}

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's control plane finalizer cloud formation stack")

		var newAdapter *adapter.CPF
		{
			c := adapter.Config{
				CustomObject:      cr,
				HostClients:       *r.hostClients,
				EncrypterBackend:  r.encrypterBackend,
				PublicRouteTables: r.publicRouteTables,
				Route53Enabled:    r.route53Enabled,
			}
			newAdapter, err = adapter.NewCPF(ctx, c)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		templateBody, err = templates.Render(key.CloudFormationHostPostTemplates(), newAdapter)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed the template of the tenant cluster's control plane finalizer cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "requesting the creation of the tenant cluster's control plane finalizer cloud formation stack")

		i := &cloudformation.CreateStackInput{
			EnableTerminationProtection: aws.Bool(key.EnableTerminationProtection),
			StackName:                   aws.String(key.MainHostPostStackName(cr)),
			Tags:                        r.getCloudFormationTags(cr),
			TemplateBody:                aws.String(templateBody),
		}

		_, err = r.hostClients.CloudFormation.CreateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "requested the creation of the tenant cluster's control plane finalizer cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "waiting for the creation of the tenant cluster's control plane finalizer cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.MainHostPreStackName(cr)),
		}

		err = r.hostClients.CloudFormation.WaitUntilStackCreateComplete(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "waited for the creation of the tenant cluster's control plane finalizer cloud formation stack")
	}

	return nil
}
