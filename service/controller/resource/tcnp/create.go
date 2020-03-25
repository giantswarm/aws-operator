package tcnp

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/controller/resource/tcnp/template"
)

const (
	capabilityNamesIAM = "CAPABILITY_NAMED_IAM"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// Ensure some preconditions are met so we have all neccessary information
	// available to manage the TCNP CF stack.
	{
		if len(cc.Spec.TenantCluster.TCNP.AvailabilityZones) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "availability zone information not yet available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		if len(cc.Status.TenantCluster.TCCP.AvailabilityZones) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "availability zone information not yet available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		if len(cc.Status.TenantCluster.TCCP.SecurityGroups) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "security group information not yet available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		if cc.Status.TenantCluster.TCCP.VPC.PeeringConnectionID == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "vpc peering connection id not yet available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's node pool cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.StackNameTCNP(&cr)),
		}

		o, err := cc.Client.TenantCluster.AWS.CloudFormation.DescribeStacks(i)
		if IsNotExists(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's node pool cloud formation stack")

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
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster's node pool cloud formation stack has stack status %#q", cloudformation.StackStatusCreateInProgress))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusUpdateInProgress {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster's node pool cloud formation stack has stack status %#q", cloudformation.StackStatusUpdateInProgress))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's node pool cloud formation stack already exists")
	}

	{
		scale, err := r.detection.ShouldScale(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}
		update, err := r.detection.ShouldUpdate(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		if scale || update {
			err = r.updateStack(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	return nil
}

func (r *Resource) createStack(ctx context.Context, cr infrastructurev1alpha2.AWSMachineDeployment) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's node pool cloud formation stack")

		params, err := newTemplateParams(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		templateBody, err = template.Render(params)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed the template of the tenant cluster's node pool cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "requesting the creation of the tenant cluster's node pool cloud formation stack")

		i := &cloudformation.CreateStackInput{
			Capabilities: []*string{
				aws.String(capabilityNamesIAM),
			},
			EnableTerminationProtection: aws.Bool(true),
			StackName:                   aws.String(key.StackNameTCNP(&cr)),
			Tags:                        r.getCloudFormationTags(cr),
			TemplateBody:                aws.String(templateBody),
		}

		_, err = cc.Client.TenantCluster.AWS.CloudFormation.CreateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "requested the creation of the tenant cluster's node pool cloud formation stack")
	}

	return nil
}

func (r *Resource) getCloudFormationTags(cr infrastructurev1alpha2.AWSMachineDeployment) []*cloudformation.Tag {
	tags := key.AWSTags(&cr, r.installationName)
	tags[key.TagStack] = key.StackTCNP
	tags[key.TagMachineDeployment] = key.MachineDeploymentID(&cr)
	return awstags.NewCloudFormation(tags)
}

func (r *Resource) updateStack(ctx context.Context, cr infrastructurev1alpha2.AWSMachineDeployment) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's node pool cloud formation stack")

		params, err := newTemplateParams(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		templateBody, err = template.Render(params)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed the template of the tenant cluster's node pool cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "requesting the update of the tenant cluster's node pool cloud formation stack")

		i := &cloudformation.UpdateStackInput{
			Capabilities: []*string{
				aws.String(capabilityNamesIAM),
			},
			StackName:    aws.String(key.StackNameTCNP(&cr)),
			TemplateBody: aws.String(templateBody),
		}
		_, err = cc.Client.TenantCluster.AWS.CloudFormation.UpdateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "requested the update of the tenant cluster's node pool cloud formation stack")
	}
	return nil
}

// minDesiredWorkers calculates appropriate minimum value to be set for ASG
// Desired value and to be used for computation of workerCountRatio.
//
// When cluster-autoscaler has scaled cluster and ASG's Desired value is higher
// than minimum number of instances allowed for that ASG, then it makes sense to
// consider Desired value as minimum number of running instances for further
// operational computations.
//
// Example:
// Initially ASG has minimum of 3 workers and maximum of 10. Due to amount of
// workload deployed on workers, cluster-autoscaler has scaled current Desired
// number of instances to 5. Therefore it makes sense to consider 5 as minimum
// number of nodes also when working on batch updates on ASG instances.
//
// Example 2:
// When end user is scaling cluster and adding restrictions to its size, it
// might be that initial ASG configuration is following:
// 		- Min: 3
//		- Max: 10
// 		- Desired: 10
//
// Now end user decides that it must be scaled down so maximum size is decreased
// to 7. When desired number of instances is temporarily bigger than maximum
// number of instances, it must be fixed to be maximum number of instances.
//
func minDesiredWorkers(minWorkers, maxWorkers, statusDesiredCapacity int) int {
	if statusDesiredCapacity > maxWorkers {
		return maxWorkers
	}

	if statusDesiredCapacity > minWorkers {
		return statusDesiredCapacity
	}

	return minWorkers
}

func newAutoScalingGroup(ctx context.Context, cr infrastructurev1alpha2.AWSMachineDeployment) (*template.ParamsMainAutoScalingGroup, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var subnets []string
	for _, az := range cc.Spec.TenantCluster.TCNP.AvailabilityZones {
		subnets = append(subnets, key.SanitizeCFResourceName(key.PrivateSubnetName(az.Name)))
	}

	minDesiredNodes := minDesiredWorkers(key.MachineDeploymentScalingMin(cr), key.MachineDeploymentScalingMax(cr), cc.Status.TenantCluster.TCNP.ASG.DesiredCapacity)

	autoScalingGroup := &template.ParamsMainAutoScalingGroup{
		AvailabilityZones: key.MachineDeploymentAvailabilityZones(cr),
		Cluster: template.ParamsMainAutoScalingGroupCluster{
			ID: key.ClusterID(&cr),
		},
		DesiredCapacity:       minDesiredNodes,
		MaxBatchSize:          workerCountRatio(minDesiredNodes, 0.3),
		MaxSize:               key.MachineDeploymentScalingMax(cr),
		MinInstancesInService: workerCountRatio(minDesiredNodes, 0.7),
		MinSize:               key.MachineDeploymentScalingMin(cr),
		Subnets:               subnets,
	}

	return autoScalingGroup, nil
}

func newIAMPolicies(ctx context.Context, cr infrastructurev1alpha2.AWSMachineDeployment) (*template.ParamsMainIAMPolicies, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var iamPolicies *template.ParamsMainIAMPolicies
	{
		iamPolicies = &template.ParamsMainIAMPolicies{
			Cluster: template.ParamsMainIAMPoliciesCluster{
				ID: key.ClusterID(&cr),
			},
			EC2ServiceDomain: key.EC2ServiceDomain(cc.Status.TenantCluster.AWS.Region),
			KMSKeyARN:        cc.Status.TenantCluster.Encryption.Key,
			NodePool: template.ParamsMainIAMPoliciesNodePool{
				ID: key.MachineDeploymentID(&cr),
			},
			RegionARN: key.RegionARN(cc.Status.TenantCluster.AWS.Region),
			S3Bucket:  key.BucketName(&cr, cc.Status.TenantCluster.AWS.AccountID),
		}
	}

	return iamPolicies, nil
}

func newLaunchConfiguration(ctx context.Context, cr infrastructurev1alpha2.AWSMachineDeployment) (*template.ParamsMainLaunchConfiguration, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	launchConfiguration := &template.ParamsMainLaunchConfiguration{
		BlockDeviceMapping: template.ParamsMainLaunchConfigurationBlockDeviceMapping{
			Docker: template.ParamsMainLaunchConfigurationBlockDeviceMappingDocker{
				Volume: template.ParamsMainLaunchConfigurationBlockDeviceMappingDockerVolume{
					Size: key.MachineDeploymentDockerVolumeSizeGB(cr),
				},
			},
			Kubelet: template.ParamsMainLaunchConfigurationBlockDeviceMappingKubelet{
				Volume: template.ParamsMainLaunchConfigurationBlockDeviceMappingKubeletVolume{
					Size: key.MachineDeploymentKubeletVolumeSizeGB(cr),
				},
			},
			Logging: template.ParamsMainLaunchConfigurationBlockDeviceMappingLogging{
				Volume: template.ParamsMainLaunchConfigurationBlockDeviceMappingLoggingVolume{
					Size: 100,
				},
			},
		},
		Instance: template.ParamsMainLaunchConfigurationInstance{
			Image:      key.ImageID(cc.Status.TenantCluster.AWS.Region),
			Monitoring: true,
			Type:       key.MachineDeploymentInstanceType(cr),
		},
		SmallCloudConfig: template.ParamsMainLaunchConfigurationSmallCloudConfig{
			S3URL: fmt.Sprintf("s3://%s/%s", key.BucketName(&cr, cc.Status.TenantCluster.AWS.AccountID), key.S3ObjectPathTCNP(&cr)),
		},
	}

	return launchConfiguration, nil
}

func newOutputs(ctx context.Context, cr infrastructurev1alpha2.AWSMachineDeployment) (*template.ParamsMainOutputs, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	outputs := &template.ParamsMainOutputs{
		DockerVolumeSizeGB: key.MachineDeploymentDockerVolumeSizeGB(cr),
		Instance: template.ParamsMainOutputsInstance{
			Image: key.ImageID(cc.Status.TenantCluster.AWS.Region),
			Type:  key.MachineDeploymentInstanceType(cr),
		},
		OperatorVersion: key.OperatorVersion(&cr),
	}

	return outputs, nil
}

func newRouteTables(ctx context.Context, cr infrastructurev1alpha2.AWSMachineDeployment) (*template.ParamsMainRouteTables, error) {
	var routeTables template.ParamsMainRouteTables

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, a := range cc.Spec.TenantCluster.TCNP.AvailabilityZones {
		r := template.ParamsMainRouteTablesListItem{
			AvailabilityZone: a.Name,
			ClusterID:        key.ClusterID(&cr),
			Name:             key.SanitizeCFResourceName(key.PrivateRouteTableName(a.Name)),
			Route: template.ParamsMainRouteTablesListItemRoute{
				Name: key.SanitizeCFResourceName(key.NATRouteName(a.Name)),
			},
			TCCP: template.ParamsMainRouteTablesListItemTCCP{
				NATGateway: template.ParamsMainRouteTablesListItemTCCPNATGateway{
					ID: a.NATGateway.ID,
				},
				VPC: template.ParamsMainRouteTablesListItemTCCPVPC{
					ID: cc.Status.TenantCluster.TCCP.VPC.ID,
				},
			},
		}
		routeTables.List = append(routeTables.List, r)
	}

	return &routeTables, nil
}

func newSecurityGroups(ctx context.Context, cr infrastructurev1alpha2.AWSMachineDeployment) (*template.ParamsMainSecurityGroups, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var nodePools []template.ParamsMainSecurityGroupsTenantClusterNodePool
	for _, ID := range cc.Spec.TenantCluster.TCNP.SecurityGroupIDs {
		np := template.ParamsMainSecurityGroupsTenantClusterNodePool{
			ID:           ID,
			ResourceName: key.SanitizeCFResourceName(ID),
		}

		nodePools = append(nodePools, np)
	}

	securityGroups := &template.ParamsMainSecurityGroups{
		ClusterID: key.ClusterID(&cr),
		ControlPlane: template.ParamsMainSecurityGroupsControlPlane{
			VPC: template.ParamsMainSecurityGroupsControlPlaneVPC{
				CIDR: cc.Status.ControlPlane.VPC.CIDR,
			},
		},
		TenantCluster: template.ParamsMainSecurityGroupsTenantCluster{
			InternalAPI: template.ParamsMainSecurityGroupsTenantClusterInternalAPI{
				ID: idFromGroups(cc.Status.TenantCluster.TCCP.SecurityGroups, key.SecurityGroupName(&cr, "internal-api")),
			},
			Master: template.ParamsMainSecurityGroupsTenantClusterMaster{
				ID: idFromGroups(cc.Status.TenantCluster.TCCP.SecurityGroups, key.SecurityGroupName(&cr, "master")),
			},
			AWSCNI: template.ParamsMainSecurityGroupsTenantClusterAWSCNI{
				ID: idFromGroups(cc.Status.TenantCluster.TCCP.SecurityGroups, key.SecurityGroupName(&cr, "aws-cni")),
			},
			NodePools: nodePools,
			VPC: template.ParamsMainSecurityGroupsTenantClusterVPC{
				ID: cc.Status.TenantCluster.TCCP.VPC.ID,
			},
		},
	}

	return securityGroups, nil
}

func newSubnets(ctx context.Context, cr infrastructurev1alpha2.AWSMachineDeployment) (*template.ParamsMainSubnets, error) {
	var subnets template.ParamsMainSubnets

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, a := range cc.Spec.TenantCluster.TCNP.AvailabilityZones {
		s := template.ParamsMainSubnetsListItem{
			AvailabilityZone: a.Name,
			CIDR:             a.Subnet.Private.CIDR.String(),
			Name:             key.SanitizeCFResourceName(key.PrivateSubnetName(a.Name)),
			RouteTable: template.ParamsMainSubnetsListItemRouteTable{
				Name: key.SanitizeCFResourceName(key.PrivateRouteTableName(a.Name)),
			},
			RouteTableAssociation: template.ParamsMainSubnetsListItemRouteTableAssociation{
				Name: key.SanitizeCFResourceName(key.PrivateSubnetRouteTableAssociationName(a.Name)),
			},
			TCCP: template.ParamsMainSubnetsListItemTCCP{
				VPC: template.ParamsMainSubnetsListItemTCCPVPC{
					ID: cc.Status.TenantCluster.TCCP.VPC.ID,
				},
			},
		}

		subnets.List = append(subnets.List, s)
	}

	return &subnets, nil
}

func newTemplateParams(ctx context.Context, cr infrastructurev1alpha2.AWSMachineDeployment) (*template.ParamsMain, error) {
	var params *template.ParamsMain
	{
		autoScalingGroup, err := newAutoScalingGroup(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		iamPolicies, err := newIAMPolicies(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		launchConfiguration, err := newLaunchConfiguration(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		outputs, err := newOutputs(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		routeTables, err := newRouteTables(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		securityGroups, err := newSecurityGroups(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		subnets, err := newSubnets(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		vpc, err := newVPC(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		params = &template.ParamsMain{
			AutoScalingGroup:    autoScalingGroup,
			IAMPolicies:         iamPolicies,
			LaunchConfiguration: launchConfiguration,
			Outputs:             outputs,
			RouteTables:         routeTables,
			SecurityGroups:      securityGroups,
			Subnets:             subnets,
			VPC:                 vpc,
		}
	}

	return params, nil
}

func newVPC(ctx context.Context, cr infrastructurev1alpha2.AWSMachineDeployment) (*template.ParamsMainVPC, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var routeTables []template.ParamsMainVPCRouteTable
	for _, a := range cc.Spec.TenantCluster.TCNP.AvailabilityZones {
		r := template.ParamsMainVPCRouteTable{
			ControlPlane: template.ParamsMainVPCRouteTableControlPlane{
				VPC: template.ParamsMainVPCRouteTableControlPlaneVPC{
					CIDR: cc.Status.ControlPlane.VPC.CIDR,
				},
			},
			Route: template.ParamsMainVPCRouteTableRoute{
				Name: key.SanitizeCFResourceName(key.VPCPeeringRouteName(a.Name)),
			},
			RouteTable: template.ParamsMainVPCRouteTableRouteTable{
				Name: key.SanitizeCFResourceName(key.PrivateRouteTableName(a.Name)),
			},
			TenantCluster: template.ParamsMainVPCRouteTableTenantCluster{
				PeeringConnectionID: cc.Status.TenantCluster.TCCP.VPC.PeeringConnectionID,
			},
		}
		routeTables = append(routeTables, r)
	}

	vpc := &template.ParamsMainVPC{
		Cluster: template.ParamsMainVPCCluster{
			ID: key.ClusterID(&cr),
		},
		Region: template.ParamsMainVPCRegion{
			ARN:  key.RegionARN(cc.Status.TenantCluster.AWS.Region),
			Name: cc.Status.TenantCluster.AWS.Region,
		},
		RouteTables: routeTables,
		TCCP: template.ParamsMainVPCTCCP{
			VPC: template.ParamsMainVPCTCCPVPC{
				ID: cc.Status.TenantCluster.TCCP.VPC.ID,
			},
		},
		TCNP: template.ParamsMainVPCTCNP{
			CIDR: key.MachineDeploymentSubnet(cr),
		},
	}

	return vpc, nil
}

func idFromGroups(groups []*ec2.SecurityGroup, name string) string {
	for _, g := range groups {
		if awstags.ValueForKey(g.Tags, "Name") == name {
			return *g.GroupId
		}
	}

	return ""
}

func workerCountRatio(workers int, ratio float32) string {
	value := float32(workers) * ratio
	rounded := int(value + 0.5)

	if rounded == 0 {
		rounded = 1
	}

	return strconv.Itoa(rounded)
}
