package tcnp

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tcnp/template"
)

const (
	capabilityNamesIAM = "CAPABILITY_NAMED_IAM"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	md, err := key.ToMachineDeployment(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// Fetch the cluster for region information and the like.
	var cl v1alpha1.Cluster
	{
		m, err := r.cmaClient.ClusterV1alpha1().Clusters(md.Namespace).Get(key.ClusterID(&md), metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		cl = *m
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's node pool cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.StackNameTCNP(&md)),
		}

		o, err := cc.Client.TenantCluster.AWS.CloudFormation.DescribeStacks(i)
		if IsNotExists(err) {
			// fall through

		} else if err != nil {
			return microerror.Mask(err)

		} else if len(o.Stacks) != 1 {
			return microerror.Maskf(executionFailedError, "expected one stack, got %d", len(o.Stacks))

		} else if *o.Stacks[0].StackStatus == cloudformation.StackStatusCreateFailed {
			return microerror.Maskf(executionFailedError, "expected successful status, got %#q", o.Stacks[0].StackStatus)

		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's node pool cloud formation stack already exists")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's node pool cloud formation stack")
	}

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's node pool cloud formation stack")

		params, err := newTemplateParams(ctx, cl, md)
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
			StackName:                   aws.String(key.StackNameTCNP(&md)),
			Tags:                        r.getCloudFormationTags(md),
			TemplateBody:                aws.String(templateBody),
		}

		_, err = cc.Client.TenantCluster.AWS.CloudFormation.CreateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "requested the creation of the tenant cluster's node pool cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "waiting for the creation of the tenant cluster's node pool cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.StackNameTCNP(&md)),
		}

		err = cc.Client.TenantCluster.AWS.CloudFormation.WaitUntilStackCreateComplete(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "waited for the creation of the tenant cluster's node pool cloud formation stack")
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

func newAutoScalingGroup(ctx context.Context, cl v1alpha1.Cluster, md v1alpha1.MachineDeployment) (*template.ParamsMainAutoScalingGroup, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var subnets []string
	for _, az := range cc.Spec.TenantCluster.TCNP.AvailabilityZones {
		subnets = append(subnets, key.SanitizeCFResourceName(key.PrivateSubnetName(az.Name)))
	}

	minDesiredNodes := minDesiredWorkers(key.WorkerScalingMin(md), key.WorkerScalingMax(md), cc.Status.TenantCluster.TCCP.ASG.DesiredCapacity)

	autoScalingGroup := &template.ParamsMainAutoScalingGroup{
		AvailabilityZones: key.WorkerAvailabilityZones(md),
		Cluster: template.ParamsMainAutoScalingGroupCluster{
			ID: key.ClusterID(&md),
		},
		DesiredCapacity:       minDesiredNodes,
		MaxBatchSize:          workerCountRatio(minDesiredNodes, 0.3),
		MaxSize:               key.WorkerScalingMax(md),
		MinInstancesInService: workerCountRatio(minDesiredNodes, 0.7),
		MinSize:               key.WorkerScalingMin(md),
		Name:                  key.MachineDeploymentASGName(&md),
		Subnets:               subnets,
	}

	return autoScalingGroup, nil
}

func newIAMPolicies(ctx context.Context, cl v1alpha1.Cluster, md v1alpha1.MachineDeployment) (*template.ParamsMainIAMPolicies, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	iamPolicies := &template.ParamsMainIAMPolicies{
		Cluster: template.ParamsMainIAMPoliciesCluster{
			ID: key.ClusterID(&md),
		},
		EC2ServiceDomain: key.EC2ServiceDomain(cl),
		KMSKeyARN:        cc.Status.TenantCluster.Encryption.Key,
		NodePool: template.ParamsMainIAMPoliciesNodePool{
			ID: key.MachineDeploymentID(&md),
		},
		RegionARN: key.RegionARN(cl),
		S3Bucket:  key.BucketName(&md, cc.Status.TenantCluster.AWSAccountID),
	}

	return iamPolicies, nil
}

func newLaunchConfiguration(ctx context.Context, cl v1alpha1.Cluster, md v1alpha1.MachineDeployment) (*template.ParamsMainLaunchConfiguration, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	launchConfiguration := &template.ParamsMainLaunchConfiguration{
		BlockDeviceMapping: template.ParamsMainLaunchConfigurationBlockDeviceMapping{
			Docker: template.ParamsMainLaunchConfigurationBlockDeviceMappingDocker{
				Volume: template.ParamsMainLaunchConfigurationBlockDeviceMappingDockerVolume{
					Size: key.WorkerDockerVolumeSizeGB(md),
				},
			},
			Logging: template.ParamsMainLaunchConfigurationBlockDeviceMappingLogging{
				Volume: template.ParamsMainLaunchConfigurationBlockDeviceMappingLoggingVolume{
					Size: 100,
				},
			},
		},
		Instance: template.ParamsMainLaunchConfigurationInstance{
			Image:      key.ImageID(cl),
			Monitoring: true,
			Type:       key.WorkerInstanceType(md),
		},
		SmallCloudConfig: template.ParamsMainLaunchConfigurationSmallCloudConfig{
			S3URL: key.SmallCloudConfigS3URL(&md, cc.Status.TenantCluster.AWSAccountID, "worker"),
		},
	}

	return launchConfiguration, nil
}

func newLifecycleHooks(ctx context.Context, cl v1alpha1.Cluster, md v1alpha1.MachineDeployment) (*template.ParamsMainLifecycleHooks, error) {
	return &template.ParamsMainLifecycleHooks{}, nil
}

func newOutputs(ctx context.Context, cl v1alpha1.Cluster, md v1alpha1.MachineDeployment) (*template.ParamsMainOutputs, error) {
	outputs := &template.ParamsMainOutputs{
		CloudConfig: template.ParamsMainOutputsCloudConfig{
			Version: key.CloudConfigVersion,
		},
		DockerVolumeSizeGB: key.WorkerDockerVolumeSizeGB(md),
		Instance: template.ParamsMainOutputsInstance{
			Image: key.ImageID(cl),
			Type:  key.WorkerInstanceType(md),
		},
		VersionBundle: template.ParamsMainOutputsVersionBundle{
			Version: key.OperatorVersion(&md),
		},
	}

	return outputs, nil
}

func newSecurityGroups(ctx context.Context, cl v1alpha1.Cluster, md v1alpha1.MachineDeployment) (*template.ParamsMainSecurityGroups, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	securityGroups := &template.ParamsMainSecurityGroups{
		ControlPlane: template.ParamsMainSecurityGroupsControlPlane{
			VPC: template.ParamsMainSecurityGroupsControlPlaneVPC{
				CIDR: cc.Status.ControlPlane.VPC.CIDR,
			},
		},
		TenantCluster: template.ParamsMainSecurityGroupsTenantCluster{
			VPC: template.ParamsMainSecurityGroupsTenantClusterVPC{
				ID: cc.Status.TenantCluster.TCCP.VPC.ID,
			},
		},
	}

	return securityGroups, nil
}

func newSubnets(ctx context.Context, cl v1alpha1.Cluster, md v1alpha1.MachineDeployment) (*template.ParamsMainSubnets, error) {
	var subnets template.ParamsMainSubnets

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	azMap := statusAZsToPublicSubnetIDs(cc.Status.TenantCluster.TCCP.AvailabilityZones)

	for _, a := range cc.Spec.TenantCluster.TCNP.AvailabilityZones {
		// Create private subnet per AZ
		s := template.ParamsMainSubnetsListItem{
			AvailabilityZone: a.Name,
			CIDR:             a.Subnet.Private.CIDR.String(),
			Name:             key.SanitizeCFResourceName(key.PrivateSubnetName(a.Name)),
			RouteTableAssociation: template.ParamsMainSubnetsListItemRouteTableAssociation{
				Name: key.SanitizeCFResourceName(key.PrivateSubnetRouteTableAssociationName(a.Name)),
			},
			TCCP: template.ParamsMainSubnetsListItemTCCP{
				Subnet: template.ParamsMainSubnetsListItemTCCPSubnet{
					ID: azMap[a.Name],
					RouteTable: template.ParamsMainSubnetsListItemTCCPSubnetRouteTable{
						Name: key.SanitizeCFResourceName(key.PublicRouteTableName(a.Name)),
					},
				},
				VPC: template.ParamsMainSubnetsListItemTCCPVPC{
					ID: cc.Status.TenantCluster.TCCP.VPC.ID,
				},
			},
		}

		subnets.List = append(subnets.List, s)
	}

	return &subnets, nil
}

func newTemplateParams(ctx context.Context, cl v1alpha1.Cluster, md v1alpha1.MachineDeployment) (*template.ParamsMain, error) {
	var params *template.ParamsMain
	{
		autoScalingGroup, err := newAutoScalingGroup(ctx, cl, md)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		iamPolicies, err := newIAMPolicies(ctx, cl, md)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		launchConfiguration, err := newLaunchConfiguration(ctx, cl, md)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		lifecycleHooks, err := newLifecycleHooks(ctx, cl, md)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		outputs, err := newOutputs(ctx, cl, md)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		securityGroups, err := newSecurityGroups(ctx, cl, md)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		subnets, err := newSubnets(ctx, cl, md)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		params = &template.ParamsMain{
			AutoScalingGroup:    autoScalingGroup,
			IAMPolicies:         iamPolicies,
			LaunchConfiguration: launchConfiguration,
			LifecycleHooks:      lifecycleHooks,
			Outputs:             outputs,
			SecurityGroups:      securityGroups,
			Subnets:             subnets,
		}
	}

	return params, nil
}

func statusAZsToPublicSubnetIDs(azs []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone) map[string]string {
	m := make(map[string]string)
	for _, az := range azs {
		m[az.Name] = az.Subnet.Public.ID
	}
	return m
}

func workerCountRatio(workers int, ratio float32) string {
	value := float32(workers) * ratio
	rounded := int(value + 0.5)

	if rounded == 0 {
		rounded = 1
	}

	return strconv.Itoa(rounded)
}
