package tcdp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/resource/tcdp/template"
)

const (
	capabilityNamesIAM = "CAPABILITY_NAMED_IAM"
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

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's data plane cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.MainHostPreStackName(cr)),
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
			r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's data plane cloud formation stack already exists")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's data plane cloud formation stack")
	}

	var templateBody string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing the template of the tenant cluster's data plane cloud formation stack")

		var params *template.ParamsMain
		{
			autoScalingGroup, err := r.newAutoScalingGroup(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}
			iamPolicies, err := r.newIAMPolicies(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}
			launchConfiguration, err := r.newLaunchConfiguration(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}
			lifecycleHooks, err := r.newLifecycleHooks(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}
			outputs, err := r.newOutputs(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}
			securityGroups, err := r.newSecurityGroups(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
			}
			subnets, err := r.newSubnets(ctx, cr)
			if err != nil {
				return microerror.Mask(err)
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

		templateBody, err = template.Render(params)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed the template of the tenant cluster's data plane cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "requesting the creation of the tenant cluster's data plane cloud formation stack")

		i := &cloudformation.CreateStackInput{
			Capabilities: []*string{
				aws.String(capabilityNamesIAM),
			},
			EnableTerminationProtection: aws.Bool(key.EnableTerminationProtection),
			StackName:                   aws.String(stackName(cr)),
			Tags:                        r.getCloudFormationTags(cr),
			TemplateBody:                aws.String(templateBody),
		}

		_, err = cc.Client.TenantCluster.AWS.CloudFormation.CreateStack(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "requested the creation of the tenant cluster's data plane cloud formation stack")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "waiting for the creation of the tenant cluster's data plane cloud formation stack")

		i := &cloudformation.DescribeStacksInput{
			StackName: aws.String(key.MainHostPreStackName(cr)),
		}

		err = cc.Client.TenantCluster.AWS.CloudFormation.WaitUntilStackCreateComplete(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "waited for the creation of the tenant cluster's data plane cloud formation stack")
	}

	return nil
}

func (r *Resource) newAutoScalingGroup(ctx context.Context, cr v1alpha1.AWSConfig) (*template.ParamsMainAutoScalingGroup, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	minDesiredNodes := minDesiredWorkers(key.ScalingMin(cr), key.ScalingMax(cr), cc.Status.TenantCluster.TCCP.ASG.DesiredCapacity)

	autoScalingGroup := &template.ParamsMainAutoScalingGroup{
		AvailabilityZones: key.StatusAvailabilityZoneNames(cr),
		Cluster: template.ParamsMainAutoScalingGroupCluster{
			ID: key.ClusterID(cr),
		},
		DesiredCapacity:       minDesiredNodes,
		MaxBatchSize:          workerCountRatio(minDesiredNodes, 0.3),
		MaxSize:               key.ScalingMax(cr),
		MinInstancesInService: workerCountRatio(minDesiredNodes, 0.7),
		MinSize:               key.ScalingMin(cr),
		Name:                  asgName(cr),
		Subnets:               key.PrivateSubnetNames(cr),
	}

	return autoScalingGroup, nil
}

func (r *Resource) newIAMPolicies(ctx context.Context, cr v1alpha1.AWSConfig) (*template.ParamsMainIAMPolicies, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	iamPolicies := &template.ParamsMainIAMPolicies{
		Cluster: template.ParamsMainIAMPoliciesCluster{
			ID: key.ClusterID(cr),
		},
		EC2ServiceDomain: key.EC2ServiceDomain(cr),
		KMSKeyARN:        cc.Status.TenantCluster.KMS.KeyARN,
		NodePool: template.ParamsMainIAMPoliciesNodePool{
			ID: nodePoolID(cr),
		},
		RegionARN: key.RegionARN(cr),
		S3Bucket:  key.BucketName(cr, cc.Status.TenantCluster.AWSAccountID),
	}

	return iamPolicies, nil
}

func (r *Resource) newLaunchConfiguration(ctx context.Context, cr v1alpha1.AWSConfig) (*template.ParamsMainLaunchConfiguration, error) {
	imageID, err := key.ImageID(cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	launchConfiguration := &template.ParamsMainLaunchConfiguration{
		BlockDeviceMapping: template.ParamsMainLaunchConfigurationBlockDeviceMapping{
			Docker: template.ParamsMainLaunchConfigurationBlockDeviceMappingDocker{
				Volume: template.ParamsMainLaunchConfigurationBlockDeviceMappingDockerVolume{
					Size: key.WorkerDockerVolumeSizeGB(cr),
				},
			},
			Logging: template.ParamsMainLaunchConfigurationBlockDeviceMappingLogging{
				Volume: template.ParamsMainLaunchConfigurationBlockDeviceMappingLoggingVolume{
					Size: 100,
				},
			},
		},
		Instance: template.ParamsMainLaunchConfigurationInstance{
			Image:      imageID,
			Monitoring: r.instanceMonitoring,
			Type:       key.WorkerInstanceType(cr),
		},
	}

	return launchConfiguration, nil
}

func (r *Resource) newLifecycleHooks(ctx context.Context, cr v1alpha1.AWSConfig) (*template.ParamsMainLifecycleHooks, error) {
	return &template.ParamsMainLifecycleHooks{}, nil
}

func (r *Resource) newOutputs(ctx context.Context, cr v1alpha1.AWSConfig) (*template.ParamsMainOutputs, error) {
	imageID, err := key.ImageID(cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	outputs := &template.ParamsMainOutputs{
		CloudConfig: template.ParamsMainOutputsCloudConfig{
			Version: key.CloudConfigVersion,
		},
		DockerVolumeSizeGB: key.WorkerDockerVolumeSizeGB(cr),
		Instance: template.ParamsMainOutputsInstance{
			Image: imageID,
			Type:  key.WorkerInstanceType(cr),
		},
		VersionBundle: template.ParamsMainOutputsVersionBundle{
			Version: key.VersionBundleVersion(cr),
		},
	}

	return outputs, nil
}

func (r *Resource) newSecurityGroups(ctx context.Context, cr v1alpha1.AWSConfig) (*template.ParamsMainSecurityGroups, error) {
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
				ID: cc.Status.TenantCluster.VPC.ID,
			},
		},
	}

	return securityGroups, nil
}

func (r *Resource) newSubnets(ctx context.Context, cr v1alpha1.AWSConfig) (*template.ParamsMainSubnets, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var subnets *template.ParamsMainSubnets

	for _, a := range key.StatusNodePoolAvailabilityZones(cr) {
		s := template.ParamsMainSubnetsListItem{
			AvailabilityZone: a.Name,
			CIDR:             a.Subnet.CIDR,
			NameSuffix:       strings.ToUpper(a.Name),
			RouteTableAssociation: template.ParamsMainSubnetsListItemRouteTableAssociation{
				NameSuffix: strings.ToUpper(a.Name),
			},
			TCCP: template.ParamsMainSubnetsListItemTCCP{
				Subnet: template.ParamsMainSubnetsListItemTCCPSubnet{
					ID: todo,
					RouteTable: template.ParamsMainSubnetsListItemTCCPSubnetRouteTable{
						ID: todo,
					},
				},
				VPC: template.ParamsMainSubnetsListItemTCCPVPC{
					ID: cc.Status.TenantCluster.VPC.ID,
				},
			},
		}

		subnets.List = append(subnets.List, s)
	}

	return subnets, nil
}

func asgName(cr v1alpha1.AWSConfig) string {
	return fmt.Sprintf("asg-%s-tcdp", key.ClusterID(cr))
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

// TODO for the tenant cluster migration we simply hard code something here.
// Once we are clear with the reconcilable types for the node pools we have to
// generate the types according to the node pools with the hardcoded ID below.
func nodePoolID(cr v1alpha1.AWSConfig) string {
	return "pb6m9"
}

func stackName(cr v1alpha1.AWSConfig) string {
	return fmt.Sprintf("cluster-%s-tcdp", key.ClusterID(cr))
}

func workerCountRatio(workers int, ratio float32) string {
	value := float32(workers) * ratio
	rounded := int(value + 0.5)

	if rounded == 0 {
		rounded = 1
	}

	return strconv.Itoa(rounded)
}
