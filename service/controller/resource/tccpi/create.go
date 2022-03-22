package tccpi

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v5/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpi/template"
)

const (
	capabilityNamesIAM = "CAPABILITY_NAMED_IAM"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.Debugf(ctx, "finding the tenant cluster's control plane initializer cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.StackNameTCCPI(&cr)),
		}

		o, err := cc.Client.ControlPlane.AWS.CloudFormation.DescribeStacks(i)
		if IsNotExists(err) {
			// fall through

		} else if err != nil {
			return microerror.Mask(err)

		} else if len(o.Stacks) != 1 {
			return microerror.Maskf(executionFailedError, "expected one stack, got %d", len(o.Stacks))

		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusCreateFailed {
			return microerror.Maskf(eventCFCreateError, "expected successful status, got %#q", *o.Stacks[0].StackStatus)
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusRollbackFailed {
			return microerror.Maskf(eventCFRollbackError, "expected successful status, got %#q", *o.Stacks[0].StackStatus)
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusUpdateRollbackFailed {
			return microerror.Maskf(eventCFUpdateRollbackError, "expected successful status, got %#q", *o.Stacks[0].StackStatus)

		} else {
			r.logger.Debugf(ctx, "found the tenant cluster's control plane initializer cloud formation stack already exists")
			r.logger.Debugf(ctx, "canceling resource")

			return nil
		}

		r.logger.Debugf(ctx, "did not find the tenant cluster's control plane initializer cloud formation stack")
	}

	var templateBody string
	{
		r.logger.Debugf(ctx, "computing the template of the tenant cluster's control plane initializer cloud formation stack")

		params, err := newTemplateParams(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		templateBody, err = template.Render(params)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "computed the template of the tenant cluster's control plane initializer cloud formation stack")
	}

	{
		r.logger.Debugf(ctx, "requesting the creation of the tenant cluster's control plane initializer cloud formation stack")

		i := &cloudformation.CreateStackInput{
			Capabilities: []*string{
				aws.String(capabilityNamesIAM),
			},
			EnableTerminationProtection: aws.Bool(true),
			StackName:                   aws.String(key.StackNameTCCPI(&cr)),
			Tags:                        r.getCloudFormationTags(cr),
			TemplateBody:                aws.String(templateBody),
		}

		_, err = cc.Client.ControlPlane.AWS.CloudFormation.CreateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "requested the creation of the tenant cluster's control plane initializer cloud formation stack")
	}

	{
		r.logger.Debugf(ctx, "waiting for the creation of the tenant cluster's control plane initializer cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.StackNameTCCPI(&cr)),
		}

		err = cc.Client.ControlPlane.AWS.CloudFormation.WaitUntilStackCreateComplete(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "waited for the creation of the tenant cluster's control plane initializer cloud formation stack")
	}

	return nil
}

func newIAMRolesParams(ctx context.Context, cr infrastructurev1alpha3.AWSCluster) (*template.ParamsMainIAMRoles, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	iamRoles := &template.ParamsMainIAMRoles{
		PeerAccessRoleName: key.RolePeerAccess(cr),
		Tenant: template.ParamsMainIAMRolesTenant{
			AWS: template.ParamsMainIAMRolesTenantAWS{
				Account: template.ParamsMainIAMRolesTenantAWSAccount{
					ID: cc.Status.TenantCluster.AWS.AccountID,
				},
			},
		},
	}

	return iamRoles, nil
}

func newTemplateParams(ctx context.Context, cr infrastructurev1alpha3.AWSCluster) (*template.ParamsMain, error) {
	var params *template.ParamsMain
	{
		iamRoles, err := newIAMRolesParams(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		params = &template.ParamsMain{
			IAMRoles: iamRoles,
		}
	}

	return params, nil
}
