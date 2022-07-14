package tccpn

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/v2/pkg/awstags"
	"github.com/giantswarm/aws-operator/v2/pkg/label"
	"github.com/giantswarm/aws-operator/v2/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v2/service/controller/key"
	"github.com/giantswarm/aws-operator/v2/service/controller/resource/tccpn/template"
	"github.com/giantswarm/aws-operator/v2/service/internal/encrypter/kms"
	"github.com/giantswarm/aws-operator/v2/service/internal/hamaster"
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

	_, err = r.encrypter.EncryptionKey(ctx, key.ClusterID(&cr))
	if kms.IsKeyNotFound(err) {
		r.logger.Debugf(ctx, "canceling resource", "reason", "encryption key not available yet")
		return nil

	} else if err != nil {
		return microerror.Mask(err)
	}

	{
		if !cc.Status.TenantCluster.S3Object.Uploaded {
			r.logger.Debugf(ctx, "s3 object not available yet")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		}

		if cc.Status.TenantCluster.TCCP.VPC.PeeringConnectionID == "" {
			r.logger.Debugf(ctx, "vpc peering connection id not available yet")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		}

		if len(cc.Status.TenantCluster.TCCP.Subnets) == 0 {
			r.logger.Debugf(ctx, "subnets not available yet")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		}

		if len(cc.Status.TenantCluster.TCCP.SecurityGroups) == 0 {
			r.logger.Debugf(ctx, "security groups not available yet")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		}

		if r.route53Enabled && (cc.Status.TenantCluster.DNS.HostedZoneID == "" || cc.Status.TenantCluster.DNS.InternalHostedZoneID == "") {
			r.logger.Debugf(ctx, "hosted zone id not available yet")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		}

		// When the TCCPN cloud formation stack is transitioning, it means it is
		// updating in most cases. We do not want to interfere with the current
		// process and stop here. We will then check on the next reconciliation loop
		// and continue eventually.
		if cc.Status.TenantCluster.TCCPN.IsTransitioning {
			r.logger.Debugf(ctx, "the tenant cluster's control plane nodes cloud formation stack is in transitioning state")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		}
	}

	{
		r.logger.Debugf(ctx, "finding the tenant cluster's control plane nodes cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.StackNameTCCPN(&cr)),
		}

		o, err := cc.Client.TenantCluster.AWS.CloudFormation.DescribeStacks(i)
		if IsNotExists(err) {
			r.logger.Debugf(ctx, "did not find the tenant cluster's control plane nodes cloud formation stack")
			err = r.createStack(ctx, cr)
			if IsNotFound(err) || hamaster.IsNotFound(err) {
				r.logger.Debugf(ctx, "not updating cloud formation stack", "reason", "CR not available yet")
				r.logger.Debugf(ctx, "canceling resource")
				return nil

			} else if IsTooManyCRsError(err) {
				r.logger.Debugf(ctx, "not updating cloud formation stack", "reason", "too many CRs found")
				r.logger.Debugf(ctx, "canceling resource")
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
			return microerror.Maskf(eventCFCreateError, "expected successful status, got %#q", *o.Stacks[0].StackStatus)
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusRollbackFailed {
			return microerror.Maskf(eventCFRollbackError, "expected successful status, got %#q", *o.Stacks[0].StackStatus)
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusUpdateRollbackFailed {
			return microerror.Maskf(eventCFUpdateRollbackError, "expected successful status, got %#q", *o.Stacks[0].StackStatus)

		} else if key.StackInProgress(*o.Stacks[0].StackStatus) {
			r.logger.Debugf(ctx, "the tenant cluster's control plane nodes cloud formation stack has stack status %#q", *o.Stacks[0].StackStatus)
			r.event.Emit(ctx, &cr, "CFInProgress", fmt.Sprintf("the tenant cluster's control plane nodes cloud formation stack has stack status %#q", *o.Stacks[0].StackStatus))
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		} else if key.StackComplete(*o.Stacks[0].StackStatus) {
			r.event.Emit(ctx, &cr, "CFCompleted", fmt.Sprintf("the tenant cluster's control plane nodes cloud formation stack has stack status %#q", *o.Stacks[0].StackStatus))
		}

		r.logger.Debugf(ctx, "found the tenant cluster's control plane nodes cloud formation stack already exists")
	}

	{
		update, err := r.detection.ShouldUpdate(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		if update {
			err = r.updateStack(ctx, cr)
			if IsNotFound(err) || hamaster.IsNotFound(err) {
				r.logger.Debugf(ctx, "not updating cloud formation stack", "reason", "CR not available yet")
				r.logger.Debugf(ctx, "canceling resource")
				return nil

			} else if IsTooManyCRsError(err) {
				r.logger.Debugf(ctx, "not updating cloud formation stack", "reason", "too many CRs found")
				r.logger.Debugf(ctx, "canceling resource")
				return nil

			} else if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	return nil
}

func (r *Resource) createStack(ctx context.Context, cr infrastructurev1alpha3.AWSControlPlane) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var templateBody string
	{
		r.logger.Debugf(ctx, "computing the template of the tenant cluster's control plane nodes cloud formation stack")

		params, err := r.newTemplateParams(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		templateBody, err = template.Render(params)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "computed the template of the tenant cluster's control plane nodes cloud formation stack")
	}

	{
		r.logger.Debugf(ctx, "requesting the creation of the tenant cluster's control plane nodes cloud formation stack")

		tags, err := r.getCloudFormationTags(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		i := &cloudformation.CreateStackInput{
			Capabilities: []*string{
				aws.String(capabilityNamesIAM),
			},
			EnableTerminationProtection: aws.Bool(true),
			StackName:                   aws.String(key.StackNameTCCPN(&cr)),
			Tags:                        tags,
			TemplateBody:                aws.String(templateBody),
		}

		_, err = cc.Client.TenantCluster.AWS.CloudFormation.CreateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "requested the creation of the tenant cluster's control plane nodes cloud formation stack")
		r.event.Emit(ctx, &cr, "CFCreateRequested", "requested the creation of the tenant cluster's control plane nodes cloud formation stack")
	}

	return nil
}

func (r *Resource) getCloudFormationTags(ctx context.Context, cr infrastructurev1alpha3.AWSControlPlane) ([]*cloudformation.Tag, error) {
	tags := key.AWSTags(&cr, r.installationName)
	if cp := key.ControlPlaneID(&cr); cp != "" {
		tags[key.TagControlPlane] = cp
	}
	tags[key.TagStack] = key.StackTCCPN

	cloudtags, err := r.cloudTags.GetTagsByCluster(ctx, key.ClusterID(&cr))
	if err != nil {
		return nil, microerror.Mask(err)
	}
	for k, v := range cloudtags {
		tags[k] = v
	}

	return awstags.NewCloudFormation(tags), nil
}

func (r *Resource) updateStack(ctx context.Context, cr infrastructurev1alpha3.AWSControlPlane) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var templateBody string
	{
		r.logger.Debugf(ctx, "computing the template of the tenant cluster's control plane nodes cloud formation stack")

		params, err := r.newTemplateParams(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		templateBody, err = template.Render(params)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "computed the template of the tenant cluster's control plane nodes cloud formation stack")
	}

	{
		r.logger.Debugf(ctx, "requesting the update of the tenant cluster's control plane nodes cloud formation stack")

		tags, err := r.getCloudFormationTags(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		i := &cloudformation.UpdateStackInput{
			Capabilities: []*string{
				aws.String(capabilityNamesIAM),
			},
			StackName:    aws.String(key.StackNameTCCPN(&cr)),
			Tags:         tags,
			TemplateBody: aws.String(templateBody),
		}

		_, err = cc.Client.TenantCluster.AWS.CloudFormation.UpdateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "requested the update of the tenant cluster's control plane nodes cloud formation stack")
		r.event.Emit(ctx, &cr, "CFUpdateRequested", "requested the update of the tenant cluster's control plane nodes cloud formation stack")
	}

	return nil
}

func (r *Resource) newAutoScalingGroup(ctx context.Context, cr infrastructurev1alpha3.AWSControlPlane) (*template.ParamsMainAutoScalingGroup, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var mappings []hamaster.Mapping
	{
		mappings, err = r.haMaster.Mapping(ctx, &cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	var haMastersEnabled bool
	{
		haMastersEnabled, err = r.haMaster.Enabled(ctx, &cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	autoScalingGroup := &template.ParamsMainAutoScalingGroup{
		HAMasters: haMastersEnabled,
	}
	for _, m := range mappings {
		dependsOn := []string{key.ControlPlaneENIResourceName(m.ID), key.ControlPlaneVolumeResourceName(m.ID)}
		// ASG for second and third master will have chain dependency on the previous one
		// to have rolling update of one ASG after the previous one.
		if m.ID == 2 || m.ID == 3 {
			dependsOn = append(dependsOn, key.ControlPlaneASGResourceName(&cr, m.ID-1))
		}
		item := template.ParamsMainAutoScalingGroupItem{
			AvailabilityZone: m.AZ,
			ClusterID:        key.ClusterID(&cr),
			DependsOn:        dependsOn,
			LaunchTemplate: template.ParamsMainAutoScalingGroupItemLaunchTemplate{
				Resource: key.ControlPlaneLaunchTemplateResourceName(&cr, m.ID),
			},
			LoadBalancers: template.ParamsMainAutoScalingGroupItemLoadBalancers{
				ApiInternalName: key.InternalELBNameAPI(&cr),
				ApiName:         key.ELBNameAPI(&cr),
				EtcdName:        key.ELBNameEtcd(&cr),
			},
			Resource: key.ControlPlaneASGResourceName(&cr, m.ID),
			SubnetID: idFromSubnets(cc.Status.TenantCluster.TCCP.Subnets, key.SanitizeCFResourceName(key.PrivateSubnetName(m.AZ))),
		}

		autoScalingGroup.List = append(autoScalingGroup.List, item)
	}

	return autoScalingGroup, nil
}

func (r *Resource) newENI(ctx context.Context, cr infrastructurev1alpha3.AWSControlPlane) (*template.ParamsMainENI, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var mappings []hamaster.Mapping
	{
		mappings, err = r.haMaster.Mapping(ctx, &cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	enis := &template.ParamsMainENI{}
	for _, m := range mappings {
		item := template.ParamsMainENIItem{
			Name:            key.ControlPlaneENIName(&cr, m.ID),
			Resource:        key.ControlPlaneENIResourceName(m.ID),
			SecurityGroupID: idFromGroups(cc.Status.TenantCluster.TCCP.SecurityGroups, key.SecurityGroupName(&cr, "master")),
			SubnetID:        idFromSubnets(cc.Status.TenantCluster.TCCP.Subnets, key.SanitizeCFResourceName(key.PrivateSubnetName(m.AZ))),
		}

		enis.List = append(enis.List, item)

	}

	return enis, nil
}

func (r *Resource) newEtcdVolume(ctx context.Context, cr infrastructurev1alpha3.AWSControlPlane) (*template.ParamsMainEtcdVolume, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var mappings []hamaster.Mapping
	{
		mappings, err = r.haMaster.Mapping(ctx, &cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	etcdVolumes := &template.ParamsMainEtcdVolume{}
	for _, m := range mappings {
		item := template.ParamsMainEtcdVolumeItem{
			AvailabilityZone: m.AZ,
			Iops:             key.ControlPlaneVolumeIops(cr),
			Name:             key.ControlPlaneVolumeName(&cr, m.ID),
			Resource:         key.ControlPlaneVolumeResourceName(m.ID),
			SnapshotID:       key.ControlPlaneVolumeSnapshotID(cc.Status.TenantCluster.MasterInstance.EtcdVolumeSnapshotID, m.ID),
			Throughput:       key.ControlPlaneVolumeThroughput(cr),
		}

		etcdVolumes.List = append(etcdVolumes.List, item)
	}

	return etcdVolumes, nil
}

func (r *Resource) newIAMPolicies(ctx context.Context, cr infrastructurev1alpha3.AWSControlPlane) (*template.ParamsMainIAMPolicies, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	ek, err := r.encrypter.EncryptionKey(ctx, key.ClusterID(&cr))
	if kms.IsKeyNotFound(err) {
		r.logger.Debugf(ctx, "canceling resource", "reason", "encryption key not found")
		return nil, nil

	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	var iamPolicies *template.ParamsMainIAMPolicies
	{
		iamPolicies = &template.ParamsMainIAMPolicies{
			AccountID:            cc.Status.TenantCluster.AWS.AccountID,
			AWSBaseDomain:        key.AWSBaseDomain(cc.Status.TenantCluster.AWS.Region),
			ClusterID:            key.ClusterID(&cr),
			EC2ServiceDomain:     key.EC2ServiceDomain(cc.Status.TenantCluster.AWS.Region),
			HostedZoneID:         cc.Status.TenantCluster.DNS.HostedZoneID,
			InternalHostedZoneID: cc.Status.TenantCluster.DNS.InternalHostedZoneID,
			KMSKeyARN:            ek,
			Region:               cc.Status.TenantCluster.AWS.Region,
			RegionARN:            key.RegionARN(cc.Status.TenantCluster.AWS.Region),
			S3Bucket:             key.BucketName(&cr, cc.Status.TenantCluster.AWS.AccountID),
			Route53Enabled:       r.route53Enabled,
		}
	}

	return iamPolicies, nil
}

func (r *Resource) newLaunchTemplate(ctx context.Context, cr infrastructurev1alpha3.AWSControlPlane) (*template.ParamsMainLaunchTemplate, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var mappings []hamaster.Mapping
	{
		mappings, err = r.haMaster.Mapping(ctx, &cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var ami string
	{
		ami, err = r.images.AMI(ctx, &cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	launchTemplate := &template.ParamsMainLaunchTemplate{}
	for _, m := range mappings {
		item := template.ParamsMainLaunchTemplateItem{
			BlockDeviceMapping: template.ParamsMainLaunchTemplateItemBlockDeviceMapping{
				Docker: template.ParamsMainLaunchTemplateItemBlockDeviceMappingDocker{
					Volume: template.ParamsMainLaunchTemplateItemBlockDeviceMappingDockerVolume{
						Size: defaultVolumeSize,
					},
				},
				Kubelet: template.ParamsMainLaunchTemplateItemBlockDeviceMappingKubelet{
					Volume: template.ParamsMainLaunchTemplateItemBlockDeviceMappingKubeletVolume{
						Size: defaultVolumeSize,
					},
				},
				Logging: template.ParamsMainLaunchTemplateItemBlockDeviceMappingLogging{
					Volume: template.ParamsMainLaunchTemplateItemBlockDeviceMappingLoggingVolume{
						Size: defaultVolumeSize,
					},
				},
			},
			Instance: template.ParamsMainLaunchTemplateItemInstance{
				Image:      ami,
				Monitoring: false,
				Type:       key.ControlPlaneInstanceType(cr),
			},
			MasterSecurityGroupID: idFromGroups(cc.Status.TenantCluster.TCCP.SecurityGroups, key.SecurityGroupName(&cr, "master")),
			Metadata: template.ParamsMainLaunchTemplateMetadata{
				HttpTokens: key.ControlPlaneMetadataV2(cr),
			},
			Name:           key.ControlPlaneLaunchTemplateName(&cr, m.ID),
			ReleaseVersion: key.ReleaseVersion(&cr),
			Resource:       key.ControlPlaneLaunchTemplateResourceName(&cr, m.ID),
			SmallCloudConfig: template.ParamsMainLaunchTemplateItemSmallCloudConfig{
				S3URL: fmt.Sprintf("s3://%s/%s", key.BucketName(&cr, cc.Status.TenantCluster.AWS.AccountID), key.S3ObjectPathTCCPN(&cr, m.ID)),
			},
		}

		launchTemplate.List = append(launchTemplate.List, item)
	}

	return launchTemplate, nil
}

func (r *Resource) newOutputs(ctx context.Context, cr infrastructurev1alpha3.AWSControlPlane) (*template.ParamsMainOutputs, error) {
	var err error

	// The reconcliation acts upon the AWSControlPlane CR, but the replicas are
	// defined in the G8sControlPlane CR. Therefore we use the HA Masters service
	// implementation to fetch the actual replicas count of the master setup.
	var rep int
	{
		rep, err = r.haMaster.Replicas(ctx, &cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	outputs := &template.ParamsMainOutputs{
		InstanceType:    key.ControlPlaneInstanceType(cr),
		MasterReplicas:  rep,
		OperatorVersion: key.OperatorVersion(&cr),
		ReleaseVersion:  key.ReleaseVersion(&cr),
	}

	return outputs, nil
}

func (r *Resource) newRecordSets(ctx context.Context, cr infrastructurev1alpha3.AWSControlPlane) (*template.ParamsMainRecordSets, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var mappings []hamaster.Mapping
	{
		mappings, err = r.haMaster.Mapping(ctx, &cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// We need to fetch the cluster CR for once because it holds the base domain
	// which we need to get the record sets right.
	var cl infrastructurev1alpha3.AWSCluster
	{
		var list infrastructurev1alpha3.AWSClusterList

		err := r.k8sClient.CtrlClient().List(
			ctx,
			&list,
			client.InNamespace(cr.GetNamespace()),
			client.MatchingLabels{label.Cluster: key.ClusterID(&cr)},
		)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if len(list.Items) == 0 {
			return nil, microerror.Mask(notFoundError)
		}
		if len(list.Items) > 1 {
			return nil, microerror.Mask(tooManyCRsError)
		}

		cl = list.Items[0]
	}

	var records []template.ParamsMainRecordSetsRecord
	for _, m := range mappings {
		item := template.ParamsMainRecordSetsRecord{
			ENI: template.ParamsMainRecordSetsRecordENI{
				Resource: key.ControlPlaneENIResourceName(m.ID),
			},
			Resource: key.ControlPlaneRecordSetsResourceName(m.ID),
			Value:    key.ControlPlaneRecordSetsRecordValue(m.ID),
		}

		records = append(records, item)
	}

	recordSets := &template.ParamsMainRecordSets{
		ClusterID:            key.ClusterID(&cr),
		InternalHostedZoneID: cc.Status.TenantCluster.DNS.InternalHostedZoneID,
		BaseDomain:           key.ClusterBaseDomain(cl),
		Records:              records,
		Route53Enabled:       r.route53Enabled,
	}

	return recordSets, nil
}

func (r *Resource) newTemplateParams(ctx context.Context, cr infrastructurev1alpha3.AWSControlPlane) (*template.ParamsMain, error) {
	var params *template.ParamsMain
	{
		autoScalingGroup, err := r.newAutoScalingGroup(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		eni, err := r.newENI(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		etcdVolume, err := r.newEtcdVolume(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		iamPolicies, err := r.newIAMPolicies(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		launchTemplate, err := r.newLaunchTemplate(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		outputs, err := r.newOutputs(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		recordSets, err := r.newRecordSets(ctx, cr)
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
			RecordSets:       recordSets,
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
