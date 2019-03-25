// +build k8srequired

package setup

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"
	"github.com/kubernetes/client-go/dynamic"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/rest"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/integration/key"
)

const (
	crNamespace = "default"
)

func EnsureTenantClusterCreated(ctx context.Context, id string, config Config, wait bool) error {
	config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating tenant cluster %#q", id))

	err := ensureAWSConfigInstalled(ctx, id, config)
	if err != nil {
		return microerror.Mask(err)
	}

	err = ensureCertConfigsInstalled(ctx, id, config)
	if err != nil {
		return microerror.Mask(err)
	}

	if wait {
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for guest cluster %#q to be ready", id))

		err = config.Guest.Initialize()
		if err != nil {
			return microerror.Mask(err)
		}

		err := config.Guest.WaitForGuestReady(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for guest cluster %#q to be ready", id))
	}

	config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created tenant cluster %#q", id))
	return nil
}

func EnsureTenantClusterDeleted(ctx context.Context, id string, config Config, wait bool) error {
	config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting tenant cluster %#q", id))

	err := config.Release.EnsureDeleted(ctx, key.AWSConfigReleaseName(id), crNotFoundCondition(ctx, config, providerv1alpha1.NewAWSConfigCRD(), crNamespace, id))
	if err != nil {
		return microerror.Mask(err)
	}

	err = config.Release.EnsureDeleted(ctx, key.CertsReleaseName(id), config.Release.Condition().SecretNotExist(ctx, "default", fmt.Sprintf("%s-api", id)))
	if err != nil {
		return microerror.Mask(err)
	}

	if wait {
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for guest cluster %#q API to be down", id))

		err := config.Guest.WaitForAPIDown()
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for guest cluster %#q API to be down", id))
	}

	config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted tenant cluster %#q", id))
	return nil
}

func crExistsCondition(ctx context.Context, config Config, crd *apiextensionsv1beta1.CustomResourceDefinition, crNamespace, crName string) release.ConditionFunc {
	return func() error {
		gvr := schema.GroupVersionResource{
			Group:    crd.Spec.Group,
			Version:  crd.Spec.Version,
			Resource: crd.Spec.Names.Plural,
		}

		var dynamicClient dynamic.Interface
		{
			var err error

			dynamicClient, err = dynamic.NewForConfig(rest.CopyConfig(config.Host.RestConfig()))
			if err != nil {
				return microerror.Mask(err)
			}
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for creation of CR %#q in namespace %#q", crName, crNamespace))

		o := func() error {
			_, err := dynamicClient.Resource(gvr).Namespace(crNamespace).Get(crName, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return microerror.Maskf(notFoundError, "CR %#q in namespace %#q", crName, crNamespace)
			} else if err != nil {
				return backoff.Permanent(microerror.Mask(err))
			}

			return nil
		}
		b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)
		n := backoff.NewNotifier(config.Logger, ctx)
		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for creation of CR %#q in namespace %#q", crName, crNamespace))
		return nil
	}
}

func crNotFoundCondition(ctx context.Context, config Config, crd *apiextensionsv1beta1.CustomResourceDefinition, crNamespace, crName string) release.ConditionFunc {
	return func() error {
		gvr := schema.GroupVersionResource{
			Group:    crd.Spec.Group,
			Version:  crd.Spec.Version,
			Resource: crd.Spec.Names.Plural,
		}

		var dynamicClient dynamic.Interface
		{
			var err error

			dynamicClient, err = dynamic.NewForConfig(rest.CopyConfig(config.Host.RestConfig()))
			if err != nil {
				return microerror.Mask(err)
			}
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for deletion of CR %#q in namespace %#q", crName, crNamespace))

		o := func() error {
			_, err := dynamicClient.Resource(gvr).Namespace(crNamespace).Get(crName, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return nil
			} else if err != nil {
				return backoff.Permanent(microerror.Mask(err))
			}

			return microerror.Maskf(stillExistsError, "CR %#q in namespace %#q", crName, crNamespace)
		}
		b := backoff.NewExponential(60*time.Minute, 5*time.Minute)
		n := backoff.NewNotifier(config.Logger, ctx)
		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for deletion of CR %#q in namespace %#q", crName, crNamespace))
		return nil
	}
}

func ensureAWSConfigInstalled(ctx context.Context, id string, config Config) error {
	vpcID, err := ensureHostVPCCreated(ctx, config)
	if err != nil {
		return microerror.Mask(err)
	}

	go func() {
		o := func() error {
			err = ensureBastionHostCreated(ctx, id, config)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}
		b := backoff.NewMaxRetries(30, 1*time.Minute)
		n := backoff.NewNotifier(config.Logger, ctx)

		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			config.Logger.LogCtx(ctx, "level", "error", "message", err.Error())
		}
	}()

	var values string
	{
		c := chartvalues.APIExtensionsAWSConfigE2EConfig{
			CommonDomain:         env.CommonDomain(),
			ClusterName:          id,
			VersionBundleVersion: env.VersionBundleVersion(),

			AWS: chartvalues.APIExtensionsAWSConfigE2EConfigAWS{
				Region:            env.AWSRegion(),
				APIHostedZone:     env.AWSAPIHostedZoneGuest(),
				IngressHostedZone: env.AWSIngressHostedZoneGuest(),
				RouteTable0:       env.AWSRouteTable0(),
				RouteTable1:       env.AWSRouteTable1(),
				VPCPeerID:         vpcID,
			},
		}

		values, err = chartvalues.NewAPIExtensionsAWSConfigE2E(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = config.Release.EnsureInstalled(ctx, key.AWSConfigReleaseName(id), release.NewStableChartInfo("apiextensions-aws-config-e2e-chart"), values, crExistsCondition(ctx, config, providerv1alpha1.NewAWSConfigCRD(), crNamespace, id))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func ensureHostVPCCreated(ctx context.Context, config Config) (string, error) {
	stackName := "host-peer-" + env.ClusterID()

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating stack %#q", stackName))

		os.Setenv("AWS_ROUTE_TABLE_0", env.AWSRouteTable0())
		os.Setenv("AWS_ROUTE_TABLE_1", env.AWSRouteTable1())
		os.Setenv("CLUSTER_NAME", env.ClusterID())
		stackInput := &cloudformation.CreateStackInput{
			StackName:        aws.String(stackName),
			TemplateBody:     aws.String(os.ExpandEnv(e2etemplates.AWSHostVPCStack)),
			TimeoutInMinutes: aws.Int64(2),
		}
		_, err := config.AWSClient.CloudFormation.CreateStack(stackInput)
		if IsStackAlreadyExists(err) {
			config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("stack %#q is already created", stackName))
		} else if err != nil {
			return "", microerror.Mask(err)
		} else {
			config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created stack %#q", stackName))
		}
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for stack %#q complete status", stackName))

		err := config.AWSClient.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		})
		if err != nil {
			return "", microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for stack %#q complete status", stackName))
	}

	var vpcID string
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding `VPCID` output in stack %#q", stackName))

		describeOutput, err := config.AWSClient.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		})
		if err != nil {
			return "", microerror.Mask(err)
		}
		for _, o := range describeOutput.Stacks[0].Outputs {
			if *o.OutputKey == "VPCID" {
				vpcID = *o.OutputValue
				break
			}
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found `VPCID` output in stack %#q", stackName))
	}

	return vpcID, nil
}

func ensureCertConfigsInstalled(ctx context.Context, id string, config Config) error {
	c := chartvalues.E2ESetupCertsConfig{
		Cluster: chartvalues.E2ESetupCertsConfigCluster{
			ID: id,
		},
		CommonDomain: env.CommonDomain(),
	}

	values, err := chartvalues.NewE2ESetupCerts(c)
	if err != nil {
		return microerror.Mask(err)
	}

	err = config.Release.EnsureInstalled(ctx, key.CertsReleaseName(id), release.NewStableChartInfo("e2esetup-certs-chart"), values, config.Release.Condition().SecretExists(ctx, "default", fmt.Sprintf("%s-api", id)))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func ensureBastionHostCreated(ctx context.Context, clusterID string, config Config) error {
	var err error

	var subnetID string
	var vpcID string
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "waiting for public subnet and vpc")

		i := &ec2.DescribeSubnetsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("tag:giantswarm.io/cluster"),
					Values: []*string{aws.String(clusterID)},
				},
				{
					Name:   aws.String("tag:aws:cloudformation:logical-id"),
					Values: []*string{aws.String("PublicSubnet")},
				},
			},
		}

		o, err := config.AWSClient.EC2.DescribeSubnets(i)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(o.Subnets) != 1 {
			return microerror.Maskf(executionFailedError, "expected one subnet, got %d", len(o.Subnets))
		}

		subnetID = *o.Subnets[0].SubnetId
		vpcID = *o.Subnets[0].VpcId

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for public subnet %#q and vpc %#q", subnetID, vpcID))
	}

	var workerSecurityGroupID string
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "waiting for worker security group")

		i := &ec2.DescribeSecurityGroupsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("tag:giantswarm.io/cluster"),
					Values: []*string{aws.String(clusterID)},
				},
				{
					Name:   aws.String("tag:aws:cloudformation:logical-id"),
					Values: []*string{aws.String("WorkerSecurityGroup")},
				},
			},
		}

		o, err := config.AWSClient.EC2.DescribeSecurityGroups(i)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(o.SecurityGroups) != 1 {
			return microerror.Maskf(executionFailedError, "expected one security group, got %d", len(o.SecurityGroups))
		}

		workerSecurityGroupID = *o.SecurityGroups[0].GroupId

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for worker security group %#q", workerSecurityGroupID))
	}

	var bastionSecurityGroupID string
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "creating bastion security group")

		i := &ec2.CreateSecurityGroupInput{
			Description: aws.String("Allow SSH access from everywhere to port 22."),
			GroupName:   aws.String(clusterID + "-bastion"),
			VpcId:       aws.String(vpcID),
		}

		o, err := config.AWSClient.EC2.CreateSecurityGroup(i)
		if err != nil {
			return microerror.Mask(err)
		}

		bastionSecurityGroupID = *o.GroupId

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created bastion security group %#q", bastionSecurityGroupID))
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "updating bastion security group to allow ssh access")

		i := &ec2.UpdateSecurityGroupRuleDescriptionsIngressInput{
			GroupId: aws.String(bastionSecurityGroupID),
			IpPermissions: []*ec2.IpPermission{
				{
					FromPort:   aws.Int64(-1),
					IpProtocol: aws.String("tcp"),
					ToPort:     aws.Int64(22),
				},
			},
		}

		_, err = config.AWSClient.EC2.UpdateSecurityGroupRuleDescriptionsIngress(i)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", "updated bastion security group to allow ssh access")
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "creating bastion instance")

		i := &ec2.RunInstancesInput{
			ImageId:      aws.String("ami-015e6cb33a709348e"),
			InstanceType: aws.String("t2.micro"),
			MaxCount:     aws.Int64(1),
			MinCount:     aws.Int64(1),
			NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
				{
					AssociatePublicIpAddress: aws.Bool(true),
					DeviceIndex:              aws.Int64(0),
					Groups: []*string{
						aws.String(bastionSecurityGroupID),
						aws.String(workerSecurityGroupID),
					},
					SubnetId: aws.String(subnetID),
				},
			},
			TagSpecifications: []*ec2.TagSpecification{
				{
					ResourceType: aws.String("instance"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(clusterID + "-bastion"),
						},
						{
							Key:   aws.String("giantswarm.io/cluster"),
							Value: aws.String(clusterID),
						},
						{
							Key:   aws.String("giantswarm.io/instance"),
							Value: aws.String("e2e-bastion"),
						},
					},
				},
			},
			UserData: aws.String(base64.StdEncoding.EncodeToString([]byte(`
				{
				  "ignition": {
				    "config": {},
				    "timeouts": {},
				    "version": "2.1.0"
				  },
				  "networkd": {},
				  "passwd": {
				    "users": [
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "xh3b4sd",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQClCCgsKl7+mQwD+giN6OEruV1ur/prpWXfyGHJyGGQkROZA3IcrpmRPWmKKXpCaW+G8lcb9DXD/K7/rNAh+4hpsfvCUs8u0mJ6u4El/8dcRTQaZUdLX8q3AZZ38gmk+yZz241x7LGd05D4H+aq9sVdtbcAepINUJyZ7p3yXTfCYwHC7QMYiuRFKMaUHY50shFhSYdD9TCEFtH2ybPi1/WOCX6gf90f6O0Ivo7tzwtYGV8ToIa2nO+CqwlIRiGqEy4/g9h1gCPDvgcLZmok74V6mH12whNdMDyJyuT8S1dLwNiKoYkvMbcUkpE0O/0LBCg+SsHVHmgnsNx9t0hUg8iR xh3b4sd"
				        ]
				      }
				    ]
				  },
				  "storage": {},
				  "systemd": {}
				}
			`))),
		}

		o, err := config.AWSClient.EC2.RunInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}

		ip := *o.Instances[0].PublicIpAddress

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created bastion instance %#q", ip))
	}

	return nil
}
