// +build k8srequired

package integration

import (
	"html/template"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
)

const (
	awsOperatorValuesFile  = "/tmp/aws-operator-values.yaml"
	awsOperatorChartValues = `Installation:
  V1:
    Name: ci-awsop
    Provider:
      AWS:
        Region: ${AWS_REGION}
    Secret:
      AWSOperator:
        IDRSAPub: ${IDRSA_PUB}
        SecretYaml: |
          service:
            aws:
              accesskey:
                id: ${AWS_ACCESS_KEY_ID}
                secret: ${AWS_SECRET_ACCESS_KEY}
                token: ${AWS_SESSION_TOKEN}
              hostaccesskey:
                id: ""
                secret: ""
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"${REGISTRY_PULL_SECRET}\"}}}"
`
	awsResourceValuesFile  = "/tmp/aws-operator-values.yaml"
	awsResourceChartValues = `commonDomain: ${COMMON_DOMAIN}
clusterName: ${CLUSTER_NAME}
clusterVersion: v_0_1_0
sshPublicKey: ${IDRSA_PUB}
versionBundleVersion: ${VERSION_BUNDLE_VERSION}
aws:
  networkCIDR: "{{.NetworkCIDR}}"
  privateSubnetCIDR: "{{.PrivateSubnetCIDR}}"
  publicSubnetCIDR: "{{.PublicSubnetCIDR}}"
  region: ${AWS_REGION}
  apiHostedZone: ${AWS_API_HOSTED_ZONE}
  ingressHostedZone: ${AWS_INGRESS_HOSTED_ZONE}
  routeTable0: ${AWS_ROUTE_TABLE_0}
  routeTable1: ${AWS_ROUTE_TABLE_1}
  vpcPeerId: ${AWS_VPC_PEER_ID}
`
)

type aWSClient struct {
	EC2 *ec2.EC2
}

func newAWSClient() aWSClient {
	awsCfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			os.Getenv("AWS_SESSION_TOKEN")),
		Region: aws.String(os.Getenv("AWS_REGION")),
	}
	s := session.New(awsCfg)
	clients := aWSClient{
		EC2: ec2.New(s),
	}

	return clients
}

var (
	f *framework
)

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	var v int
	var err error
	f, err = newFramework()
	if err != nil {
		log.Printf("unexpected error: %v\n", err)
		os.Exit(1)
	}

	if err := f.setUp(); err != nil {
		log.Printf("unexpected error: %v\n", err)
		v = 1
	}

	if err := operatorSetup(); err != nil {
		log.Printf("unexpected error: %v\n", err)
		v = 1
	}

	if v == 0 {
		v = m.Run()
	}

	f.deleteGuestCluster()
	operatorTearDown()
	f.tearDown()

	os.Exit(v)
}

func TestPodsResolveNames(t *testing.T) {
}

func operatorSetup() error {
	if err := f.installCertOperator(); err != nil {
		return microerror.Mask(err)
	}

	if err := f.installCertResource(); err != nil {
		return microerror.Mask(err)
	}

	if err := f.installAwsOperator(); err != nil {
		return microerror.Mask(err)
	}

	if err := writeAWSResourceValues(); err != nil {
		return microerror.Maskf(err, "unexpected error writing aws-resource-lab values file")
	}

	if err := runCmd("helm registry install quay.io/giantswarm/aws-resource-lab-chart:stable -- -n aws-resource-lab --values " + awsOperatorValuesFile); err != nil {
		return microerror.Maskf(err, "unexpected error installing aws-resource-lab chart: %v")
	}

	logEntry := "cluster '${CLUSTER_NAME}' processed"
	if os.Getenv("VERSION_BUNDLE_VERSION") == "0.2.0" {
		logEntry = "creating AWS cloudformation stack: created"
	}

	operatorPodName, err := f.podName("giantswarm", "app=aws-operator")
	if err != nil {
		return microerror.Maskf(err, "unexpected error getting operator pod name: %v")
	}

	if err := f.waitForPodLog("giantswarm", logEntry, operatorPodName); err != nil {
		return microerror.Maskf(err, "unexpected error waiting for guest cluster installed: %v")
	}

	if err := f.initGuestClientset(); err != nil {
		return microerror.Maskf(err, "unexpected error initializing guest clientset")
	}

	if err := f.waitForAPIUp(); err != nil {
		return microerror.Maskf(err, "unexpected error waiting for API up")
	}

	return nil
}

func operatorTearDown() {
	runCmd("helm delete cert-resource-lab --purge")
	runCmd("helm delete cert-operator --purge")
	runCmd("helm delete aws-resource-lab --purge")
	runCmd("helm delete aws-operator --purge")
}

func writeAWSResourceValues() error {
	awsResourceChartValuesEnv := os.ExpandEnv(awsResourceChartValues)

	tmpl, err := template.New("awsResource").Parse(awsResourceChartValuesEnv)
	if err != nil {
		return microerror.Mask(err)
	}

	f, err := os.Create(awsResourceValuesFile)
	if err != nil {
		return microerror.Mask(err)
	}
	defer f.Close()

	awsClient := newAWSClient()
	vpc, err := newAWSVPCBlock(awsClient)
	if err != nil {
		return microerror.Mask(err)
	}

	err = tmpl.Execute(f, vpc)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
