package cleanuptcnpiamroles

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v2/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v2/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var policies []string
	{
		r.logger.Debugf(ctx, "finding all policies")

		i := &iam.ListAttachedRolePoliciesInput{
			RoleName: aws.String(key.MachineDeploymentNodeRole(cr)),
		}

		o, err := cc.Client.TenantCluster.AWS.IAM.ListAttachedRolePolicies(i)
		if IsNotFound(err) {
			r.logger.Debugf(ctx, "no attached policies")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		for _, p := range o.AttachedPolicies {
			policies = append(policies, *p.PolicyArn)
		}

		r.logger.Debugf(ctx, "found %d policies", len(policies))
	}

	for _, p := range policies {
		r.logger.Debugf(ctx, "detaching policy %s", p)

		i := &iam.DetachRolePolicyInput{
			PolicyArn: aws.String(p),
			RoleName:  aws.String(key.MachineDeploymentNodeRole(cr)),
		}

		_, err := cc.Client.TenantCluster.AWS.IAM.DetachRolePolicy(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "detached policy %s", p)
	}

	return nil
}
