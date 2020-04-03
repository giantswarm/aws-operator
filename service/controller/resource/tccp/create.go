package tccp

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/internal/adapter"
	"github.com/giantswarm/aws-operator/service/controller/internal/ebs"
	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/internal/templates"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	// when aws operator starts it needs to find CP VPC information, so we have to
	// cancel the resource in case the information is not yet available.
	{
		if cc.Status.ControlPlane.VPC.ID == "" || cc.Status.ControlPlane.VPC.CIDR == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the control plane VPC info is not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
	}

	// When a tenant cluster is created, the CPI resource creates a peer role and
	// with it an ARN for it. As long as the peer role ARN is not present, we have
	// to cancel the resource to prevent further TCCP resource actions.
	{
		if cc.Status.ControlPlane.PeerRole.ARN == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster's control plane peer role arn is not yet set up")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
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

		if key.StatusNetworkCIDR(cr) == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's control plane network cidr")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's control plane network cidr")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's control plane cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.MainGuestStackName(cr)),
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
			MasterInstanceResourceName: key.MasterInstanceResourceName(cr),
			DockerVolumeResourceName:   key.DockerVolumeResourceName(cr),
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
			EnableTerminationProtection: aws.Bool(key.EnableTerminationProtection),
			Parameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String(versionBundleVersionParameterKey),
					ParameterValue: aws.String(key.VersionBundleVersion(cr)),
				},
			},
			StackName:    aws.String(key.MainGuestStackName(cr)),
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
					ParameterValue: aws.String(key.VersionBundleVersion(cr)),
				},
			},
			StackName:    aws.String(key.MainGuestStackName(cr)),
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
	tags := key.ClusterTags(cr, r.installationName)
	return awstags.NewCloudFormation(tags)
}

func (r *Resource) newTemplateBody(ctx context.Context, cr v1alpha1.AWSConfig, tp templateParams) (string, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var osVersion string
	{
		releaseVersion, ok := cr.Labels[label.ReleaseVersion]
		if !ok {
			return "", microerror.Maskf(executionFailedError, "release version label not found")
		}
		releaseName := fmt.Sprintf("v%s", releaseVersion)
		release, err := r.g8sClient.ReleaseV1alpha1().Releases().Get(releaseName, metav1.GetOptions{})
		if err != nil {
			return "", microerror.Mask(err)
		}
		for _, component := range release.Spec.Components {
			if component.Name == "containerlinux" {
				osVersion = component.Version
				break
			}
		}
		if osVersion == "" {
			return "", microerror.Maskf(executionFailedError, "containerlinux component version not found on release")
		}
	}

	im, err := key.ImageID(cr, osVersion)
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
			ControlPlaneVPCID:               cc.Status.ControlPlane.VPC.ID,
			CustomObject:                    cr,
			EncrypterBackend:                r.encrypterBackend,
			InstallationName:                r.installationName,
			PublicRouteTables:               r.publicRouteTables,
			Route53Enabled:                  r.route53Enabled,
			StackState: adapter.StackState{
				Name: key.MainGuestStackName(cr),

				DockerVolumeResourceName:   tp.DockerVolumeResourceName,
				MasterImageID:              im,
				MasterInstanceResourceName: tp.MasterInstanceResourceName,
				MasterInstanceType:         key.MasterInstanceType(cr),
				MasterCloudConfigVersion:   key.CloudConfigVersion,
				MasterInstanceMonitoring:   r.instanceMonitoring,

				WorkerCloudConfigVersion: key.CloudConfigVersion,
				WorkerDesired:            cc.Status.TenantCluster.TCCP.ASG.DesiredCapacity,
				WorkerDockerVolumeSizeGB: key.WorkerDockerVolumeSizeGB(cr),
				// TODO: https://github.com/giantswarm/giantswarm/issues/4105#issuecomment-421772917
				// TODO: for now we use same value as for DockerVolumeSizeFromNode, when we have kubelet size in spec we should use that.
				WorkerKubeletVolumeSizeGB: key.WorkerDockerVolumeSizeGB(cr),
				WorkerImageID:             im,
				WorkerInstanceMonitoring:  r.instanceMonitoring,
				WorkerInstanceType:        key.WorkerInstanceType(cr),
				WorkerMax:                 cc.Status.TenantCluster.TCCP.ASG.MaxSize,
				WorkerMin:                 cc.Status.TenantCluster.TCCP.ASG.MinSize,

				VersionBundleVersion: key.VersionBundleVersion(cr),
			},
			TenantClusterAccountID: cc.Status.TenantCluster.AWSAccountID,
		}

		if r.encrypterBackend == encrypter.KMSBackend {
			c.TenantClusterKMSKeyARN = cc.Status.TenantCluster.Encryption.Key
		}

		a, err := adapter.NewGuest(c)
		if err != nil {
			return "", microerror.Mask(err)
		}

		templateBody, err = templates.Render(key.CloudFormationGuestTemplates(), a)
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
		MasterInstanceResourceName: key.MasterInstanceResourceName(cr),
		DockerVolumeResourceName:   key.DockerVolumeResourceName(cr),
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
