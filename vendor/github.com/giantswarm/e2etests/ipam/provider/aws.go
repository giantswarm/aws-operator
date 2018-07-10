package provider

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	awsclient "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

// apiextensionsAWSConfigE2EChartValues is modified version of
// e2etemplates.ApiextensionsAWSConfigE2EChartValues for IPAM test scenario.
// Major difference is lack of subnet fields and need to replace clusterName
// format string value. ClusterName is not based on env variable because there
// would be too high risk for unintentional disruption of other tests if
// ${CLUSTER_NAME} would be dynamically changed on the fly.
const apiextensionsAWSConfigE2EChartValues = `commonDomain: {{ .CommonDomain }}
clusterName: %s
clusterVersion: v_0_1_0
sshPublicKey: {{ .SSHPublicKey}}
versionBundleVersion: {{ .VersionBundleVersion }}
aws:
  region: ${ .AWSRegion }
  apiHostedZone: {{ .AWSAPIHostedZoneGuest }}
  ingressHostedZone: {{ .AWSIngressHostedZone }}
  routeTable0: {{ .AWSRouteTable0 }}
  routeTable1: {{ .AWSRouteTable1 }}
  vpcPeerId: {{ .AWSVPCPeerID }}
`

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

	chartValuesConfig   ChartValuesConfig
	chartValuesTemplate *template.Template
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
			AWSVPCPeerID:              getValue(config.AWSVPCPeerID, "AWS_VPC_PEER_ID"),
			CommonDomain:              getValue(config.CommonDomain, "COMMON_DOMAIN"),
			SSHPublicKey:              getValue(config.SSHPublicKey, "IDRSA_PUB"),
			VersionBundleVersion:      getValue(config.VersionBundleVersion, "VERSION_BUNDLE_VERSION"),
		},
		chartValuesTemplate: template.Must(template.New("awsConfigChartValues").Parse(apiextensionsAWSConfigE2EChartValues)),
	}

	return a, nil
}

func (aws *AWS) CreateCluster(clusterName string) error {
	if clusterName == "" {
		return microerror.Maskf(invalidConfigError, "clusterName must not be empty")
	}

	deploymentName := awsConfigDeploymentName(clusterName)

	o := func() error {
		// NOTE we ignore errors here because we cannot get really useful error
		// handling done. This here should anyway only be a quick fix until we use
		// the helm client lib. Then error handling will be better.
		framework.HelmCmd(fmt.Sprintf("delete --purge %s", deploymentName))

		f, err := ioutil.TempFile("/tmp", deploymentName)
		if err != nil {
			return microerror.Mask(err)
		}
		defer func() {
			// Close & remove created tempfile.
			fName := f.Name()
			err = f.Close()
			if err != nil {
				// XXX: Tempfile leak.
				aws.logger.Log("level", "error", "message", fmt.Sprintf("failed to close & remove tempfile '%s'", fName))
				return
			}
			err = os.Remove(fName)
			if err != nil {
				// XXX: Tempfile leak.
				aws.logger.Log("level", "error", "message", fmt.Sprintf("failed to close & remove tempfile '%s'", fName))
				return
			}
		}()

		chartValuesConfig := aws.chartValuesConfig
		chartValuesConfig.ClusterName = clusterName
		err = aws.chartValuesTemplate.Execute(f, chartValuesConfig)
		if err != nil {
			return microerror.Mask(err)
		}

		err = framework.HelmCmd("registry install quay.io/giantswarm/apiextensions-aws-config-e2e-chart:stable -- -n aws-config-e2e --values " + f.Name())
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	b := framework.NewExponentialBackoff(framework.ShortMaxWait, framework.ShortMaxInterval)
	n := func(err error, delay time.Duration) {
		log.Println("level", "debug", "message", err.Error())
	}

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (aws *AWS) DeleteCluster(clusterName string) {
	deploymentName := awsConfigDeploymentName(clusterName)
	// NOTE we ignore errors here because we cannot get really useful error
	// handling done. This here should anyway only be a quick fix until we use
	// the helm client lib. Then error handling will be better.
	framework.HelmCmd(fmt.Sprintf("delete --purge %s", deploymentName))
}

func awsConfigDeploymentName(clusterName string) string {
	return fmt.Sprintf("aws-config-e2e-%s", clusterName)
}

// getValue returns val if not empty, otherwise os.Getenv(envName).
func getValue(val, envName string) string {
	if val == "" {
		return os.Getenv(envName)
	}
	return val
}
