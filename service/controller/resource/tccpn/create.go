package tccpn

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
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
			if err != nil {
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
			if err != nil {
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

	clusterCr, err := r.g8sClient.InfrastructureV1alpha2().AWSClusters(cr.Namespace).Get(cr.Annotations[key.TagCluster], metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	masterCount := 1

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's control plane nodes cloud formation stack")

		params, err := newTemplateParams(ctx, cr, r.route53Enabled, masterCount, key.ClusterBaseDomain(*clusterCr))
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
	clusterCr, err := r.g8sClient.InfrastructureV1alpha2().AWSClusters(cr.Namespace).Get(cr.Annotations[key.TagCluster], metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	masterCount := 0

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's control plane nodes cloud formation stack")

		params, err := newTemplateParams(ctx, cr, r.route53Enabled, masterCount, key.ClusterBaseDomain(*clusterCr))
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

func newAutoScalingGroup(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane, masterCount int) (*template.ParamsMainAutoScalingGroup, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var resourceNames []string
	{
		for i := 0; i < masterCount; i++ {
			resourceNames = append(resourceNames, key.ControlPlaneASGResourceName(i))
		}
	}

	autoScalingGroup := &template.ParamsMainAutoScalingGroup{
		AvailabilityZone: key.ControlPlaneAvailabilityZones(cr)[0],
		ClusterID:        key.ClusterID(&cr),
		LoadBalancers: template.ParamsMainAutoScalingGroupLoadBalancers{
			ApiInternalName: key.InternalELBNameAPI(&cr),
			ApiName:         key.ELBNameAPI(&cr),
			EtcdName:        key.ELBNameEtcd(&cr),
		},
		ResourceNames: resourceNames,
		SubnetID:      idFromSubnets(cc.Status.TenantCluster.TCCP.Subnets, key.SanitizeCFResourceName(key.PrivateSubnetName(key.ControlPlaneAvailabilityZones(cr)[0]))),
	}

	return autoScalingGroup, nil
}

func newENI(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane, masterCount int) (*template.ParamsMainENI, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var masterSubnets []net.IPNet
	{
		zones := cc.Spec.TenantCluster.TCCP.AvailabilityZones
		for _, az := range zones {
			masterSubnets = append(masterSubnets, az.Subnet.Private.CIDR)
		}
	}

	cpSecuritygroupID := idFromGroups(cc.Status.TenantCluster.TCCP.SecurityGroups, key.SecurityGroupName(&cr, "master"))

	var enis []template.ParamsMainENISpec
	{
		for i := 0; i < masterCount; i++ {
			e := template.ParamsMainENISpec{
				IpAddress:       key.ControlPlaneENIIpAddress(masterSubnets[i]),
				Name:            key.ControlPlaneENIName(&cr, i),
				SecurityGroupID: cpSecuritygroupID,
				SubnetID:        idFromSubnets(cc.Status.TenantCluster.TCCP.Subnets, key.SanitizeCFResourceName(key.PrivateSubnetName(key.ControlPlaneAvailabilityZones(cr)[i]))),
				ResourceName:    key.ControlPlaneENIResourceName(i),
			}
			enis = append(enis, e)
		}
	}
	eni := &template.ParamsMainENI{
		ENIs: enis,
	}

	return eni, nil
}

func newEtcdVolume(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane, masterCount int) (*template.ParamsMainEtcdVolume, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	var volumes []template.ParamsMainEtcdVolumeEtcdVolumeSpec
	{
		snapshotID := cc.Status.TenantCluster.MasterInstance.EtcdVolumeSnapshotID
		for i := 0; i < masterCount; i++ {
			if i != 0 {
				// this is weird hack but snapshotID does not make sense for all volumes, since it would not work
				snapshotID = ""
			}

			v := template.ParamsMainEtcdVolumeEtcdVolumeSpec{
				AvailabilityZone: key.ControlPlaneAvailabilityZones(cr)[i],
				Name:             key.ControlPlaneVolumeNameEtcd(&cr, i),
				SnapshotID:       snapshotID,
				ResourceName:     key.ControlPlaneVolumeResourceName(i),
			}
			volumes = append(volumes, v)
		}
	}
	etcdVolume := &template.ParamsMainEtcdVolume{
		Volumes: volumes,
	}

	return etcdVolume, nil
}

func newIAMPolicies(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane, route53Enabled bool) (*template.ParamsMainIAMPolicies, error) {
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
			Route53Enabled:       route53Enabled,
		}
	}

	return iamPolicies, nil
}

func newLaunchTemplate(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane, masterCount int) (*template.ParamsMainLaunchTemplate, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var smallCloudConfigs []template.ParamsMainLaunchTemplateSmallCloudConfig
	{
		for i := 0; i < masterCount; i++ {
			scc := template.ParamsMainLaunchTemplateSmallCloudConfig{
				S3URL: fmt.Sprintf("s3://%s/%s", key.BucketName(&cr, cc.Status.TenantCluster.AWS.AccountID), key.S3ObjectPathTCCPN(&cr, i)),
			}
			smallCloudConfigs = append(smallCloudConfigs, scc)
		}
	}

	launchTemplate := &template.ParamsMainLaunchTemplate{
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
		SmallCloudConfigs:     smallCloudConfigs,
		ResourceName:          key.ControlPlaneLaunchTemplateName(&cr, 0),
	}

	return launchTemplate, nil
}

func newRecordSets(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane, route53Enabled bool, masterCount int, baseDomain string) (*template.ParamsMainRecordSets, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var recordSets *template.ParamsMainRecordSets
	{
		var records []template.ParamsMainRecordSetsRecords
		{
			for i := 0; i < masterCount; i++ {
				record := template.ParamsMainRecordSetsRecords{
					Value:           key.ControlPlaneRecordSetsRecordValue(i),
					ResourceName:    key.ControlPlaneRecordSetsResourceName(i),
					ENIResourceName: key.ControlPlaneENIResourceName(i),
				}
				records = append(records, record)
			}
		}

		recordSets = &template.ParamsMainRecordSets{
			ClusterID:      key.ClusterID(&cr),
			HostedZoneID:   cc.Status.TenantCluster.DNS.HostedZoneID,
			BaseDomain:     baseDomain,
			Records:        records,
			Route53Enabled: route53Enabled,
		}
	}

	return recordSets, nil
}

func newOutputs(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane) (*template.ParamsMainOutputs, error) {
	outputs := &template.ParamsMainOutputs{
		InstanceType:    key.ControlPlaneInstanceType(cr),
		OperatorVersion: key.OperatorVersion(&cr),
	}

	return outputs, nil
}

func newTemplateParams(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane, route53Enabled bool, masterCount int, baseDomain string) (*template.ParamsMain, error) {
	var params *template.ParamsMain
	{
		autoScalingGroup, err := newAutoScalingGroup(ctx, cr, masterCount)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		eni, err := newENI(ctx, cr, masterCount)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		etcdVolume, err := newEtcdVolume(ctx, cr, masterCount)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		iamPolicies, err := newIAMPolicies(ctx, cr, route53Enabled)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		launchTemplate, err := newLaunchTemplate(ctx, cr, masterCount)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		recordSets, err := newRecordSets(ctx, cr, route53Enabled, masterCount, baseDomain)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		outputs, err := newOutputs(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		params = &template.ParamsMain{
			AutoScalingGroup: autoScalingGroup,
			ENI:              eni,
			EtcdVolume:       etcdVolume,
			IAMPolicies:      iamPolicies,
			LaunchTemplate:   launchTemplate,
			RecordSets:       recordSets,
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
