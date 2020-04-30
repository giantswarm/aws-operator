package tccpn

import (
	"context"
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/internal/hamaster"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/tccpn/template"
)

const (
	capabilityNamesIAM = "CAPABILITY_NAMED_IAM"
	defaultVolumeSize  = 100
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToControlPlane(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		if cc.Status.TenantCluster.Encryption.Key == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "encryption key not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		if cc.Status.TenantCluster.MasterInstance.EtcdVolumeSnapshotID == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "etcd volume snapshot id not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		if cc.Status.TenantCluster.TCCP.VPC.PeeringConnectionID == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "vpc peering connection id not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		if len(cc.Status.TenantCluster.TCCP.Subnets) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "subnets not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		if len(cc.Status.TenantCluster.TCCP.SecurityGroups) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "security groups not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		if r.route53Enabled {
			if cc.Status.TenantCluster.DNS.HostedZoneID == "" || cc.Status.TenantCluster.DNS.InternalHostedZoneID == "" {
				r.logger.LogCtx(ctx, "level", "debug", "message", "hosted zone id not available yet")
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

				return nil
			}
		}

		// When the TCCPN cloud formation stack is transitioning, it means it is
		// updating in most cases. We do not want to interfere with the current
		// process and stop here. We will then check on the next reconciliation loop
		// and continue eventually.
		if cc.Status.TenantCluster.TCCPN.IsTransitioning {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster's control plane nodes cloud formation stack is in transitioning state")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's control plane nodes cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.StackNameTCCPN(&cr)),
		}

		o, err := cc.Client.TenantCluster.AWS.CloudFormation.DescribeStacks(i)
		if IsNotExists(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's control plane nodes cloud formation stack")
			err = r.createStack(ctx, cr)
			if hamaster.IsNotFound(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "not creating cloud formation stack", "reason", "control plane CR not available yet")
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				return nil

			} else if err != nil {
				return microerror.Mask(err)
			}

			return nil

		} else if err != nil {
			return microerror.Mask(err)

		} else if len(o.Stacks) != 1 {
			return microerror.Maskf(executionFailedError, "expected one stack, got %d", len(o.Stacks))

		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusCreateFailed {
			return microerror.Maskf(executionFailedError, "expected successful status, got %#q", *o.Stacks[0].StackStatus)

		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusCreateInProgress {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster's control plane nodes cloud formation stack has stack status %#q", cloudformation.StackStatusCreateInProgress))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusUpdateInProgress {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster's control plane nodes cloud formation stack has stack status %#q", cloudformation.StackStatusUpdateInProgress))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's control plane nodes cloud formation stack already exists")
	}

	{
		update, err := r.detection.ShouldUpdate(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		if update {
			err = r.updateStack(ctx, cr)
			if hamaster.IsNotFound(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "not updating cloud formation stack", "reason", "control plane CR not available yet")
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				return nil

			} else if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	return nil
}

func (r *Resource) createStack(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's control plane nodes cloud formation stack")

		params, err := r.newTemplateParams(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		templateBody, err = template.Render(params)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed the template of the tenant cluster's control plane nodes cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "requesting the creation of the tenant cluster's control plane nodes cloud formation stack")

		i := &cloudformation.CreateStackInput{
			Capabilities: []*string{
				aws.String(capabilityNamesIAM),
			},
			EnableTerminationProtection: aws.Bool(true),
			StackName:                   aws.String(key.StackNameTCCPN(&cr)),
			Tags:                        r.getCloudFormationTags(cr),
			TemplateBody:                aws.String(templateBody),
		}

		_, err = cc.Client.TenantCluster.AWS.CloudFormation.CreateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "requested the creation of the tenant cluster's control plane nodes cloud formation stack")
	}

	return nil
}

func (r *Resource) getCloudFormationTags(cr infrastructurev1alpha2.AWSControlPlane) []*cloudformation.Tag {
	tags := key.AWSTags(&cr, r.installationName)
	tags[key.TagControlPlane] = key.ControlPlaneID(&cr)
	tags[key.TagStack] = key.StackTCCPN
	return awstags.NewCloudFormation(tags)
}

func (r *Resource) updateStack(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's control plane nodes cloud formation stack")

		params, err := r.newTemplateParams(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		templateBody, err = template.Render(params)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed the template of the tenant cluster's control plane nodes cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "requesting the update of the tenant cluster's control plane nodes cloud formation stack")

		i := &cloudformation.UpdateStackInput{
			Capabilities: []*string{
				aws.String(capabilityNamesIAM),
			},
			StackName:    aws.String(key.StackNameTCCPN(&cr)),
			TemplateBody: aws.String(templateBody),
		}

		_, err = cc.Client.TenantCluster.AWS.CloudFormation.UpdateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "requested the update of the tenant cluster's control plane nodes cloud formation stack")
	}

	return nil
}

func (r *Resource) newAutoScalingGroup(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane, haMasterEnabled bool) (*template.ParamsMainAutoScalingGroup, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	autoScalingGroup := &template.ParamsMainAutoScalingGroup{
		AvailabilityZone: key.ControlPlaneAvailabilityZones(cr)[0],
		ClusterID:        key.ClusterID(&cr),
		LoadBalancers: template.ParamsMainAutoScalingGroupLoadBalancers{
			ApiInternalName: key.InternalELBNameAPI(&cr),
			ApiName:         key.ELBNameAPI(&cr),
			EtcdName:        key.ELBNameEtcd(&cr),
		},
		SubnetID: idFromSubnets(cc.Status.TenantCluster.TCCP.Subnets, key.SanitizeCFResourceName(key.PrivateSubnetName(key.ControlPlaneAvailabilityZones(cr)[0]))),
	}

	return autoScalingGroup, nil
}

func (r *Resource) newENI(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) (*template.ParamsMainENI, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var masterSubnets []net.IPNet
	{
		zones := cc.Spec.TenantCluster.TCCP.AvailabilityZones
		for _, az := range zones {
			if az.Name != key.ControlPlaneAvailabilityZones(cr)[0] {
				continue
			}
			masterSubnets = append(masterSubnets, az.Subnet.Private.CIDR)
		}
	}

	eni := &template.ParamsMainENI{
		IpAddress:       key.ControlPlaneENIIpAddress(masterSubnets[0]),
		Name:            key.ControlPlaneENIName(&cr, 0),
		SecurityGroupID: idFromGroups(cc.Status.TenantCluster.TCCP.SecurityGroups, key.SecurityGroupName(&cr, "master")),
		SubnetID:        idFromSubnets(cc.Status.TenantCluster.TCCP.Subnets, key.SanitizeCFResourceName(key.PrivateSubnetName(key.ControlPlaneAvailabilityZones(cr)[0]))),
	}
	return eni, nil
}

func (r *Resource) newEtcdVolume(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane, haMasterEnabled bool) (*template.ParamsMainEtcdVolume, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	etcdVolume := &template.ParamsMainEtcdVolume{
		AvailabilityZone: key.ControlPlaneAvailabilityZones(cr)[0],
		Name:             key.ControlPlaneVolumeNameEtcd(&cr, 0),
		SnapshotID:       cc.Status.TenantCluster.MasterInstance.EtcdVolumeSnapshotID,
	}

	return etcdVolume, nil
}

func (r *Resource) newIAMPolicies(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) (*template.ParamsMainIAMPolicies, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var iamPolicies *template.ParamsMainIAMPolicies
	{
		iamPolicies = &template.ParamsMainIAMPolicies{
			ClusterID:            key.ClusterID(&cr),
			EC2ServiceDomain:     key.EC2ServiceDomain(cc.Status.TenantCluster.AWS.Region),
			HostedZoneID:         cc.Status.TenantCluster.DNS.HostedZoneID,
			InternalHostedZoneID: cc.Status.TenantCluster.DNS.InternalHostedZoneID,
			KMSKeyARN:            cc.Status.TenantCluster.Encryption.Key,
			RegionARN:            key.RegionARN(cc.Status.TenantCluster.AWS.Region),
			S3Bucket:             key.BucketName(&cr, cc.Status.TenantCluster.AWS.AccountID),
			Route53Enabled:       r.route53Enabled,
		}
	}

	return iamPolicies, nil
}

func (r *Resource) newLaunchTemplate(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane, haMasterEnabled bool) (*template.ParamsMainLaunchTemplate, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var ids []int
	if haMasterEnabled {
		ids = []int{1, 2, 3}
	} else {
		ids = []int{0}
	}

	var launchTemplates []template.ParamsMainLaunchTemplateItem
	for _, id := range ids {
		item := template.ParamsMainLaunchTemplateItem{
			BlockDeviceMapping: template.ParamsMainLaunchTemplateBlockDeviceMapping{
				Docker: template.ParamsMainLaunchTemplateBlockDeviceMappingDocker{
					Volume: template.ParamsMainLaunchTemplateBlockDeviceMappingDockerVolume{
						Size: defaultVolumeSize,
					},
				},
				Kubelet: template.ParamsMainLaunchTemplateBlockDeviceMappingKubelet{
					Volume: template.ParamsMainLaunchTemplateBlockDeviceMappingKubeletVolume{
						Size: defaultVolumeSize,
					},
				},
				Logging: template.ParamsMainLaunchTemplateBlockDeviceMappingLogging{
					Volume: template.ParamsMainLaunchTemplateBlockDeviceMappingLoggingVolume{
						Size: defaultVolumeSize,
					},
				},
			},
			Instance: template.ParamsMainLaunchTemplateInstance{
				Image:      key.ImageID(cc.Status.TenantCluster.AWS.Region),
				Monitoring: false,
				Type:       key.ControlPlaneInstanceType(cr),
			},
			MasterSecurityGroupID: idFromGroups(cc.Status.TenantCluster.TCCP.SecurityGroups, key.SecurityGroupName(&cr, "master")),
			SmallCloudConfig: template.ParamsMainLaunchTemplateSmallCloudConfig{
				S3URL: fmt.Sprintf("s3://%s/%s", key.BucketName(&cr, cc.Status.TenantCluster.AWS.AccountID), key.S3ObjectPathTCCPN(&cr, id)),
			},
			ResourceName: key.ControlPlaneLaunchTemplateName(&cr, 0),
		}

		launchTemplates = append(launchTemplates, item)
	}

	launchTemplate := &template.ParamsMainLaunchTemplate{
		List: launchTemplates,
	}

	return launchTemplate, nil
}

func (r *Resource) newOutputs(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) (*template.ParamsMainOutputs, error) {
	outputs := &template.ParamsMainOutputs{
		InstanceType:    key.ControlPlaneInstanceType(cr),
		OperatorVersion: key.OperatorVersion(&cr),
	}

	return outputs, nil
}

func (r *Resource) newTemplateParams(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) (*template.ParamsMain, error) {
	var err error

	// We need to determine if we want to generate certificates for a Tenant
	// Cluster with a HA Master setup.
	var haMasterEnabled bool
	{
		haMasterEnabled, err = r.haMaster.Enabled(ctx, key.ClusterID(&cr))
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var params *template.ParamsMain
	{
		autoScalingGroup, err := r.newAutoScalingGroup(ctx, cr, haMasterEnabled)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		eni, err := r.newENI(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		etcdVolume, err := r.newEtcdVolume(ctx, cr, haMasterEnabled)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		iamPolicies, err := r.newIAMPolicies(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		launchTemplate, err := r.newLaunchTemplate(ctx, cr, haMasterEnabled)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		outputs, err := r.newOutputs(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		params = &template.ParamsMain{
			AutoScalingGroup: autoScalingGroup,
			ENI:              eni,
			EtcdVolume:       etcdVolume,
			IAMPolicies:      iamPolicies,
			LaunchTemplate:   launchTemplate,
			Outputs:          outputs,
		}
	}

	return params, nil
}

func idFromGroups(groups []*ec2.SecurityGroup, name string) string {
	for _, g := range groups {
		if awstags.ValueForKey(g.Tags, "Name") == name {
			return *g.GroupId
		}
	}

	return ""
}

func idFromSubnets(subnets []*ec2.Subnet, name string) string {
	for _, s := range subnets {
		if awstags.ValueForKey(s.Tags, "Name") == name {
			return *s.SubnetId
		}
	}

	return ""
}
