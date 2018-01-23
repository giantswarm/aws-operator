// +build k8srequired

package integration

import (
	"fmt"
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
	c aWSClient
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

	c = newAWSClient()

	if err := f.SetUp(); err != nil {
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

	f.DeleteGuestCluster()
	operatorTearDown()
	f.TearDown()

	os.Exit(v)
}

func TestGuestReadyAfterMasterReboot(t *testing.T) {
	log.Println("getting master ID")
	describeInput := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(fmt.Sprintf("%s-master", os.Getenv("CLUSTER_NAME"))),
				},
			},
			&ec2.Filter{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("running"),
				},
			},
		},
	}
	res, err := c.EC2.DescribeInstances(describeInput)
	if err != nil {
		t.Errorf("unexpected error getting master id %v", err)
	}
	if len(res.Reservations) != 1 {
		t.Errorf("unexpected number of reservations %d", len(res.Reservations))
	}
	if len(res.Reservations[0].Instances) != 1 {
		t.Errorf("unexpected number of instances %d", len(res.Reservations[0].Instances))
	}
	masterID := res.Reservations[0].Instances[0].InstanceId

	log.Println("rebooting master")
	rebootInput := &ec2.RebootInstancesInput{
		InstanceIds: []*string{
			masterID,
		},
	}
	_, err = c.EC2.RebootInstances(rebootInput)
	if err != nil {
		t.Errorf("unexpected error rebooting  master %v", err)
	}

	if err := f.WaitForAPIDown(); err != nil {
		t.Errorf("unexpected error waiting for master shutting down %v", err)
	}

	if err := f.WaitForGuestReady(); err != nil {
		t.Errorf("unexpected error waiting for guest cluster ready, %v", err)
	}
}

func operatorSetup() error {
	if err := f.InstallCertOperator(); err != nil {
		return microerror.Mask(err)
	}

	if err := f.InstallCertResource(); err != nil {
		return microerror.Mask(err)
	}

	if err := f.InstallAwsOperator(); err != nil {
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

	operatorPodName, err := f.PodName("giantswarm", "app=aws-operator")
	if err != nil {
		return microerror.Maskf(err, "unexpected error getting operator pod name: %v")
	}

	if err := f.WaitForPodLog("giantswarm", logEntry, operatorPodName); err != nil {
		return microerror.Maskf(err, "unexpected error waiting for guest cluster installed: %v")
	}

	if err := f.WaitForGuestReady(); err != nil {
		return microerror.Maskf(err, "unexpected error waiting for guest cluster ready")
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

	vpc, err := newAWSVPCBlock(c)
	if err != nil {
		return microerror.Mask(err)
	}

	err = tmpl.Execute(f, vpc)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
