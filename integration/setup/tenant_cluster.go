// +build k8srequired

package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	crNamespace = "default"
)

func EnsureTenantClusterCreated(ctx context.Context, id string, config Config) error {
	err := InstallAWSConfig(ctx, id, config)
	if err != nil {
		return microerror.Mask(err)
	}

	err = InstallCertConfigs(ctx, id, config)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "waiting for guest cluster to be ready")

		err = config.Guest.Initialize()
		if err != nil {
			return microerror.Mask(err)
		}

		err = config.Guest.Setup()
		if err != nil {
			return microerror.Mask(err)
		}

		err := config.Guest.WaitForGuestReady()
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", "waited for guest cluster to be ready")
	}

	return nil
}

func EnsureTenantClusterDeleted(ctx context.Context, id string, config Config) error {
	err := config.Release.EnsureDeleted(ctx, id, CRNotExistsCondition(ctx, id, config))
	if err != nil {
		return microerror.Mask(err)
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "waiting for guest cluster API to be down")

		err := config.Guest.WaitForAPIDown()
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", "waited for guest cluster API to be down")
	}

	return nil
}

func crExistsCondition(ctx context.Context, id string, config Config) release.ConditionFunc {
	return func() error {
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for creation of CR %#q in namespace %#q", id, crNamespace))

		o := func() error {
			_, err := config.Host.G8sClient().ProviderV1alpha1().AWSConfigs(crNamespace).Get(id, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return microerror.Maskf(notFoundError, "CR %#q in namespace %#q", id, crNamespace)
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

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for creation of CR %#q in namespace %#q", id, crNamespace))
		return nil
	}
}

func CRNotExistsCondition(ctx context.Context, id string, config Config) release.ConditionFunc {
	return func() error {
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for deletion of CR %#q in namespace %#q", id, crNamespace))

		o := func() error {
			_, err := config.Host.G8sClient().ProviderV1alpha1().AWSConfigs(crNamespace).Get(id, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return nil
			} else if err != nil {
				return backoff.Permanent(microerror.Mask(err))
			}

			return microerror.Maskf(stillExistsError, "CR %#q in namespace %#q", id, crNamespace)
		}
		b := backoff.NewExponential(60*time.Minute, 5*time.Minute)
		n := backoff.NewNotifier(config.Logger, ctx)
		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for deletion of CR %#q in namespace %#q", id, crNamespace))
		return nil
	}
}

func ensureHostVPC(ctx context.Context, config Config) (string, error) {
	var err error

	var template string
	{
		c := e2etemplates.AWSHostPeerStackConfig{
			Stack: e2etemplates.AWSHostPeerStackConfigStack{
				Name: env.HostPeerStackName(),
			},
			RouteTable0: e2etemplates.AWSHostPeerStackConfigRouteTable0{
				Name: env.AWSRouteTable0(),
			},
			RouteTable1: e2etemplates.AWSHostPeerStackConfigRouteTable1{
				Name: env.AWSRouteTable1(),
			},
		}

		template, err = e2etemplates.NewAWSHostPeerStack(c)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating stack %#q", env.HostPeerStackName()))

		stackInput := &cloudformation.CreateStackInput{
			StackName:        aws.String(env.HostPeerStackName()),
			TemplateBody:     aws.String(template),
			TimeoutInMinutes: aws.Int64(2),
		}
		_, err := config.AWSClient.CloudFormation.CreateStack(stackInput)
		if IsStackAlreadyExists(err) {
			config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("stack %#q is already created", env.HostPeerStackName()))
		} else if err != nil {
			return "", microerror.Mask(err)
		} else {
			config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created stack %#q", env.HostPeerStackName()))
		}
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for stack %#q complete status", env.HostPeerStackName()))

		err := config.AWSClient.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
			StackName: aws.String(env.HostPeerStackName()),
		})
		if err != nil {
			return "", microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for stack %#q complete status", env.HostPeerStackName()))
	}

	var vpcID string
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding `VPCID` output in stack %#q", env.HostPeerStackName()))

		describeOutput, err := config.AWSClient.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String(env.HostPeerStackName()),
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

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found `VPCID` output in stack %#q", env.HostPeerStackName()))
	}

	return vpcID, nil
}

func InstallAWSConfig(ctx context.Context, id string, config Config) error {
	vpcID, err := ensureHostVPC(ctx, config)
	if err != nil {
		return microerror.Mask(err)
	}

	var values string
	{
		c := chartvalues.APIExtensionsAWSConfigE2EConfig{
			CommonDomain:         env.CommonDomain(),
			ClusterName:          id,
			SSHPublicKey:         env.IDRSAPub(),
			VersionBundleVersion: env.VersionBundleVersion(),

			AWS: chartvalues.APIExtensionsAWSConfigE2EConfigAWS{
				APIHostedZone:     env.AWSAPIHostedZoneGuest(),
				IngressHostedZone: env.AWSIngressHostedZoneGuest(),
				NetworkCIDR:       env.NetworkCIDR(),
				PrivateSubnetCIDR: env.PrivateSubnetCIDR(),
				PublicSubnetCIDR:  env.PublicSubnetCIDR(),
				Region:            env.AWSRegion(),
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

	err = config.Release.Install(ctx, fmt.Sprintf("e2esetup-awsconfig-%s", id), release.NewStableChartInfo("apiextensions-aws-config-e2e-chart"), values, crExistsCondition(ctx, id, config))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func InstallCertConfigs(ctx context.Context, id string, config Config) error {
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

	err = config.Release.Install(ctx, fmt.Sprintf("e2esetup-certs-%s", id), release.NewStableChartInfo("e2esetup-certs-chart"), values, config.Release.Condition().SecretExists(ctx, "default", fmt.Sprintf("%s-api", id)))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
