package tccp

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/adapter"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/ebs"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/legacykey"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/templates"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := legacykey.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// When the TCCP cloud formation stack is transitioning, it means it is
	// updating in most cases. We do not want to interfere with the current
	// process and stop here. We will then check on the next reconciliation loop
	// and continue eventually.
	{
		if cc.Status.TenantCluster.TCCP.IsTransitioning {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster's control plane cloud formation stack is in transitioning state")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
	}

	// The IPAM resource is executed before the CloudFormation resource in order
	// to allocate a free IP range for the tenant subnet. This CIDR is put into
	// the CR status. In case it is missing, the IPAM resource did not yet
	// allocate it and the CloudFormation resource cannot proceed. We cancel here
	// and wait for the CIDR to be available in the CR status.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's control plane network cidr")

		if legacykey.StatusNetworkCIDR(cr) == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's control plane network cidr")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's control plane network cidr")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's control plane cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(legacykey.StackNameTCCP(cr)),
		}

		o, err := cc.Client.TenantCluster.AWS.CloudFormation.DescribeStacks(i)
		if IsNotExists(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's control plane cloud formation stack")

			err = r.createStack(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil

		} else if err != nil {
			return microerror.Mask(err)

		} else if len(o.Stacks) != 1 {
			return microerror.Maskf(executionFailedError, "expected one stack, got %d", len(o.Stacks))

		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusCreateFailed {
			return microerror.Maskf(executionFailedError, "expected successful status, got %#q", o.Stacks[0].StackStatus)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's control plane cloud formation stack")
	}

	{
		update, err := r.detection.ShouldUpdate(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		if update {
			err = r.updateStack(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}
	}

	{
		scale, err := r.detection.ShouldScale(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		if scale {
			err = r.scaleStack(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}
	}

	return nil
}

func (r *Resource) createStack(ctx context.Context, cr v1alpha1.AWSConfig) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if r.encrypterBackend == encrypter.VaultBackend {
		err = r.encrypterRoleManager.EnsureCreatedAuthorizedIAMRoles(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var templateBody string
	{
		tp := templateParams{
			MasterInstanceResourceName: legacykey.MasterInstanceResourceName(cr),
			DockerVolumeResourceName:   legacykey.DockerVolumeResourceName(cr),
		}

		templateBody, err = r.newTemplateBody(ctx, cr, tp)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "requesting the creation of the tenant cluster's control plane cloud formation stack")

		i := &cloudformation.CreateStackInput{
			// CAPABILITY_NAMED_IAM is required for creating worker policy IAM roles.
			Capabilities: []*string{
				aws.String(namedIAMCapability),
			},
			EnableTerminationProtection: aws.Bool(legacykey.EnableTerminationProtection),
			Parameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String(versionBundleVersionParameterKey),
					ParameterValue: aws.String(legacykey.VersionBundleVersion(cr)),
				},
			},
			StackName:    aws.String(legacykey.StackNameTCCP(cr)),
			Tags:         r.getCloudFormationTags(cr),
			TemplateBody: aws.String(templateBody),
		}

		_, err = cc.Client.TenantCluster.AWS.CloudFormation.CreateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "requested the creation of the tenant cluster's control plane cloud formation stack")
	}

	return nil
}

func (r *Resource) detachVolumes(ctx context.Context, cr v1alpha1.AWSConfig) error {
	cc, err := controllercontext.FromContext(ctx)
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

	return nil
}

func (r *Resource) ensureStack(ctx context.Context, cr v1alpha1.AWSConfig, templateBody string) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring the tenant cluster's control plane cloud formation stack")

		i := &cloudformation.UpdateStackInput{
			Capabilities: []*string{
				// CAPABILITY_NAMED_IAM is required for updating worker policy IAM
				// roles.
				aws.String(namedIAMCapability),
			},
			Parameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String(versionBundleVersionParameterKey),
					ParameterValue: aws.String(legacykey.VersionBundleVersion(cr)),
				},
			},
			StackName:    aws.String(legacykey.StackNameTCCP(cr)),
			TemplateBody: aws.String(templateBody),
		}

		_, err = cc.Client.TenantCluster.AWS.CloudFormation.UpdateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "ensured the tenant cluster's control plane cloud formation stack")
	}

	return nil
}

func (r *Resource) getCloudFormationTags(cr v1alpha1.AWSConfig) []*cloudformation.Tag {
	tags := legacykey.ClusterTags(cr, r.installationName)
	return awstags.NewCloudFormation(tags)
}

func (r *Resource) newTemplateBody(ctx context.Context, cr v1alpha1.AWSConfig, tp templateParams) (string, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var templateBody string
	{
		c := adapter.Config{
			APIWhitelist:                    r.apiWhiteList,
			ControlPlaneAccountID:           cc.Status.ControlPlane.AWSAccountID,
			ControlPlaneNATGatewayAddresses: cc.Status.ControlPlane.NATGateway.Addresses,
			ControlPlanePeerRoleARN:         cc.Status.ControlPlane.PeerRole.ARN,
			ControlPlaneVPCCidr:             cc.Status.ControlPlane.VPC.CIDR,
			CustomObject:                    cr,
			EncrypterBackend:                r.encrypterBackend,
			InstallationName:                r.installationName,
			PublicRouteTables:               r.publicRouteTables,
			Route53Enabled:                  r.route53Enabled,
			StackState: adapter.StackState{
				Name: legacykey.StackNameTCCP(cr),

				DockerVolumeResourceName:   tp.DockerVolumeResourceName,
				MasterImageID:              legacykey.ImageID(cr),
				MasterInstanceResourceName: tp.MasterInstanceResourceName,
				MasterInstanceType:         legacykey.MasterInstanceType(cr),
				MasterCloudConfigVersion:   legacykey.CloudConfigVersion,
				MasterInstanceMonitoring:   r.instanceMonitoring,

				WorkerCloudConfigVersion: legacykey.CloudConfigVersion,
				WorkerDesired:            cc.Status.TenantCluster.TCCP.ASG.DesiredCapacity,
				WorkerDockerVolumeSizeGB: legacykey.WorkerDockerVolumeSizeGB(cr),
				// TODO: https://github.com/giantswarm/giantswarm/issues/4105#issuecomment-421772917
				// TODO: for now we use same value as for DockerVolumeSizeFromNode, when we have kubelet size in spec we should use that.
				WorkerKubeletVolumeSizeGB: legacykey.WorkerDockerVolumeSizeGB(cr),
				WorkerImageID:             legacykey.ImageID(cr),
				WorkerInstanceMonitoring:  r.instanceMonitoring,
				WorkerInstanceType:        legacykey.WorkerInstanceType(cr),
				WorkerMax:                 cc.Status.TenantCluster.TCCP.ASG.MaxSize,
				WorkerMin:                 cc.Status.TenantCluster.TCCP.ASG.MinSize,

				VersionBundleVersion: legacykey.VersionBundleVersion(cr),
			},
			TenantClusterAccountID: cc.Status.TenantCluster.AWSAccountID,
			TenantClusterKMSKeyARN: cc.Status.TenantCluster.Encryption.Key,
		}

		a, err := adapter.NewGuest(c)
		if err != nil {
			return "", microerror.Mask(err)
		}

		templateBody, err = templates.Render(legacykey.CloudFormationGuestTemplates(), a)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	return templateBody, nil
}

func (r *Resource) scaleStack(ctx context.Context, cr v1alpha1.AWSConfig) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	tp := templateParams{
		MasterInstanceResourceName: cc.Status.TenantCluster.MasterInstance.ResourceName,
		DockerVolumeResourceName:   cc.Status.TenantCluster.MasterInstance.DockerVolumeResourceName,
	}

	templateBody, err := r.newTemplateBody(ctx, cr, tp)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.ensureStack(ctx, cr, templateBody)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) updateStack(ctx context.Context, cr v1alpha1.AWSConfig) error {
	tp := templateParams{
		MasterInstanceResourceName: legacykey.MasterInstanceResourceName(cr),
		DockerVolumeResourceName:   legacykey.DockerVolumeResourceName(cr),
	}

	templateBody, err := r.newTemplateBody(ctx, cr, tp)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.detachVolumes(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.terminateMasterInstance(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.ensureStack(ctx, cr, templateBody)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
