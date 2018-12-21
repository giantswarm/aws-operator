// +build k8srequired

package setup

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
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
		// In client-go@10.12.x it will be:
		//
		//	gvr := metav1.GroupVersionResource{
		//		Group:    crd.Spec.Group,
		//		Version:  crd.Spec.Version,
		//		Resource: crd.Spec.Names.Plural,
		//	}
		//
		resource := &metav1.APIResource{
			Name:       crd.Spec.Names.Plural,
			Namespaced: crd.Spec.Scope == "Namespaced",
		}
		gv := &schema.GroupVersion{
			Group:   crd.Spec.Group,
			Version: crd.Spec.Version,
		}

		var dynamicClient *dynamic.Client
		{
			var err error

			c := config.Host.RestConfig()
			configShallowCopy := *c
			configShallowCopy.APIPath = "/apis"
			configShallowCopy.GroupVersion = gv
			if configShallowCopy.UserAgent == "" {
				configShallowCopy.UserAgent = rest.DefaultKubernetesUserAgent()
			}

			// In client-go@10.12.x it will be:
			//
			//	dynamicClient, err := dynamic.NewForConfig(config.Host.RestConfig())
			//
			dynamicClient, err = dynamic.NewClient(&configShallowCopy)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for creation of CR %#q in namespace %#q", crName, crNamespace))

		o := func() error {
			// In client-go@10.12.x it will be:
			//
			//	_, err := dynamicClient.Reosurce(gvr).Namespace(crNamespace).Get(crName, metav1.GetOptions{})
			//
			_, err := dynamicClient.Resource(resource, crNamespace).Get(crName, metav1.GetOptions{})
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
		// In client-go@10.12.x it will be:
		//
		//	gvr := metav1.GroupVersionResource{
		//		Group:    crd.Spec.Group,
		//		Version:  crd.Spec.Version,
		//		Resource: crd.Spec.Names.Plural,
		//	}
		//
		resource := &metav1.APIResource{
			Name:       crd.Spec.Names.Plural,
			Namespaced: crd.Spec.Scope == "Namespaced",
		}
		gv := &schema.GroupVersion{
			Group:   crd.Spec.Group,
			Version: crd.Spec.Version,
		}

		var dynamicClient *dynamic.Client
		{
			var err error

			c := config.Host.RestConfig()
			configShallowCopy := *c
			configShallowCopy.APIPath = "/apis"
			configShallowCopy.GroupVersion = gv
			if configShallowCopy.UserAgent == "" {
				configShallowCopy.UserAgent = rest.DefaultKubernetesUserAgent()
			}

			// In client-go@10.12.x it will be:
			//
			//	dynamicClient, err := dynamic.NewForConfig(config.Host.RestConfig())
			//
			dynamicClient, err = dynamic.NewClient(&configShallowCopy)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for deletion of CR %#q in namespace %#q", crName, crNamespace))

		o := func() error {
			// In client-go@10.12.x it will be:
			//
			//	_, err := dynamicClient.Reosurce(gvr).Namespace(crNamespace).Get(crName, metav1.GetOptions{})
			//
			_, err := dynamicClient.Resource(resource, crNamespace).Get(crName, metav1.GetOptions{})
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

	var values string
	{
		c := chartvalues.APIExtensionsAWSConfigE2EConfig{
			CommonDomain:         env.CommonDomain(),
			ClusterName:          id,
			SSHPublicKey:         env.IDRSAPub(),
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
