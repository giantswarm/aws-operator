package provider

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	awsclient "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AWSConfig struct {
	AWSClient     *awsclient.Client
	HostFramework *framework.Host
	Logger        micrologger.Logger

	ChartValuesConfig
}

type AWS struct {
	awsClient     *awsclient.Client
	hostFramework *framework.Host
	logger        micrologger.Logger

	chartValuesConfig ChartValuesConfig
	hostClusterName   string
}

type ChartValuesConfig struct {
	AWSAPIHostedZoneGuest     string
	AWSIngressHostedZoneGuest string
	AWSRegion                 string
	AWSRouteTable0            string
	AWSRouteTable1            string
	AWSVPCPeerID              string
	ClusterName               string
	CommonDomain              string
	SSHPublicKey              string
	VersionBundleVersion      string
}

func NewAWS(config AWSConfig) (*AWS, error) {
	if config.AWSClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.AWSClient must not be empty", config)
	}
	if config.HostFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostFramework must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	a := &AWS{
		awsClient:     config.AWSClient,
		hostFramework: config.HostFramework,
		logger:        config.Logger,

		chartValuesConfig: ChartValuesConfig{
			AWSAPIHostedZoneGuest:     getValue(config.AWSAPIHostedZoneGuest, "AWS_API_HOSTED_ZONE_GUEST"),
			AWSIngressHostedZoneGuest: getValue(config.AWSIngressHostedZoneGuest, "AWS_INGRESS_HOSTED_ZONE_GUEST"),
			AWSRegion:                 getValue(config.AWSRegion, "AWS_REGION"),
			AWSRouteTable0:            getValue(config.AWSRouteTable0, "AWS_ROUTE_TABLE_0"),
			AWSRouteTable1:            getValue(config.AWSRouteTable1, "AWS_ROUTE_TABLE_1"),
			CommonDomain:              getValue(config.CommonDomain, "COMMON_DOMAIN"),
			SSHPublicKey:              getValue(config.SSHPublicKey, "IDRSA_PUB"),
			VersionBundleVersion:      getValue(config.VersionBundleVersion, "VERSION_BUNDLE_VERSION"),
		},
		hostClusterName: getValue(config.ClusterName, "CLUSTER_NAME"),
	}

	return a, nil
}

func (aws *AWS) CreateCluster(clusterName string) error {
	if clusterName == "" {
		return microerror.Maskf(invalidConfigError, "clusterName must not be empty")
	}

	hostAWSConfig, err := aws.hostFramework.AWSCluster(aws.hostClusterName)
	if err != nil {
		return microerror.Mask(err)
	}

	err = aws.installCertResources(clusterName)
	if err != nil {
		return microerror.Mask(err)
	}

	deploymentName := awsConfigDeploymentName(clusterName)

	o := func() error {
		// NOTE we ignore errors here because we cannot get really useful error
		// handling done. This here should anyway only be a quick fix until we use
		// the helm client lib. Then error handling will be better.
		err := framework.HelmCmd(fmt.Sprintf("delete --purge %s", deploymentName))
		if err != nil {
			return microerror.Mask(err)
		}

		f, err := ioutil.TempFile("/tmp", deploymentName)
		if err != nil {
			return microerror.Mask(err)
		}
		defer func() {
			// Close & remove created tempfile.
			fName := f.Name()
			err = f.Close()
			if err != nil {
				// XXX: Tempfile cannot be removed due to error. It must be
				// removed manually.
				aws.logger.Log("level", "error", "message", fmt.Sprintf("failed to close & remove tempfile '%s'", fName))
				return
			}
			err = os.Remove(fName)
			if err != nil {
				// XXX: Tempfile cannot be removed due to error. It must be
				// removed manually.
				aws.logger.Log("level", "error", "message", fmt.Sprintf("failed to close & remove tempfile '%s'", fName))
				return
			}
		}()

		{
			c := chartvalues.APIExtensionsAWSConfigE2EConfig{
				CommonDomain:         aws.chartValuesConfig.CommonDomain,
				ClusterName:          clusterName,
				SSHPublicKey:         aws.chartValuesConfig.SSHPublicKey,
				VersionBundleVersion: aws.chartValuesConfig.VersionBundleVersion,

				AWS: chartvalues.APIExtensionsAWSConfigE2EConfigAWS{
					Region:            aws.chartValuesConfig.AWSRegion,
					APIHostedZone:     aws.chartValuesConfig.AWSAPIHostedZoneGuest,
					IngressHostedZone: aws.chartValuesConfig.AWSIngressHostedZoneGuest,
					RouteTable0:       clusterName + "_0",
					RouteTable1:       clusterName + "_1",
					VPCPeerID:         hostAWSConfig.Spec.AWS.VPC.PeerID,
				},
			}

			values, err := chartvalues.NewAPIExtensionsAWSConfigE2E(c)
			if err != nil {
				return microerror.Mask(err)
			}

			_, err = f.WriteString(values)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		err = framework.HelmCmd(fmt.Sprintf("registry install quay.io/giantswarm/apiextensions-aws-config-e2e-chart:stable -- -n %s --values %s", awsConfigDeploymentName(clusterName), f.Name()))
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	b := backoff.NewConstant(framework.ShortMaxWait, framework.ShortMaxInterval)
	n := func(err error, delay time.Duration) {
		log.Println("level", "debug", "message", err.Error())
	}

	err = backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (aws *AWS) DeleteCluster(clusterName string) error {
	deploymentName := awsConfigDeploymentName(clusterName)
	err := framework.HelmCmd(fmt.Sprintf("delete --purge %s", deploymentName))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (aws *AWS) installCertResources(clusterName string) error {
	{
		aws.logger.Log("level", "debug", "message", "installing cert resource chart")

		o := func() error {
			err := framework.HelmCmd(fmt.Sprintf("delete --purge %s", certDeploymentName(clusterName)))
			if err != nil {
				return microerror.Mask(err)
			}

			cmdStr := fmt.Sprintf("registry install quay.io/giantswarm/apiextensions-cert-config-e2e-chart:stable -- -n %s --set commonDomain=${COMMON_DOMAIN} --set clusterName=%s", certDeploymentName(clusterName), clusterName)
			err = framework.HelmCmd(cmdStr)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}
		b := backoff.NewExponential(framework.ShortMaxWait, framework.ShortMaxInterval)
		n := backoff.NewNotifier(aws.logger, context.Background())
		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}

		aws.logger.Log("level", "debug", "message", "installed cert resource chart")
	}

	{
		aws.logger.Log("level", "debug", "message", "waiting for k8s secret to be there")

		o := func() error {
			n := fmt.Sprintf("%s-api", clusterName)
			_, err := aws.hostFramework.K8sClient().CoreV1().Secrets("default").Get(n, metav1.GetOptions{})
			if err != nil {
				return microerror.Maskf(waitError, "k8s secret is still missing")
			}

			return nil
		}
		b := backoff.NewExponential(framework.ShortMaxWait, framework.ShortMaxInterval)
		n := func(err error, delay time.Duration) {
			aws.logger.Log("level", "debug", "message", err.Error())
		}

		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}

		aws.logger.Log("level", "debug", "message", "k8s secret is there")
	}

	return nil
}

func awsConfigDeploymentName(clusterName string) string {
	return fmt.Sprintf("aws-config-e2e-%s", clusterName)
}

func certDeploymentName(clusterName string) string {
	return fmt.Sprintf("%s-cert-config-e2e", clusterName)
}

// getValue returns val if not empty, otherwise os.Getenv(envName).
func getValue(val, envName string) string {
	if val == "" {
		return os.Getenv(envName)
	}
	return val
}
