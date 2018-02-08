// +build k8srequired

package integration

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/giantswarm/aws-operator/service"
	"github.com/giantswarm/aws-operator/service/awsconfig/v2/key"
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
  networkCIDR: "10.12.0.0/24"
  privateSubnetCIDR: "10.12.0.0/25"
  publicSubnetCIDR: "10.12.0.128/25"
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
	CF  *cloudformation.CloudFormation
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
		CF:  cloudformation.New(s),
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

	version, err := GetVersionBundleVersion(service.NewVersionBundles(), os.Getenv("TESTED_VERSION"))
	if err != nil {
		log.Printf("Unexpected error getting version bundle version %v", err)
		os.Exit(1)
	}
	if version == "" {
		log.Printf("No version bundle version for TESTED_VERSION %q", os.Getenv("TESTED_VERSION"))
		os.Exit(0)
	}
	os.Setenv("VERSION_BUNDLE_VERSION", version)

	c = newAWSClient()

	if err := createHostPeerVPC(); err != nil {
		log.Printf("unexpected error: %v\n", err)
		os.Exit(1)
	}

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
	deleteHostPeerVPC()

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

func TestWorkersScaling(t *testing.T) {
	currentWorkers, err := numberOfWorkers(os.Getenv("CLUSTER_NAME"))
	if err != nil {
		t.Fatalf("unexpected error getting number of workers %v", err)
	}
	currentMasters, err := numberOfMasters(os.Getenv("CLUSTER_NAME"))
	if err != nil {
		t.Fatalf("unexpected error getting number of masters %v", err)
	}

	// increase number of workers
	expectedWorkers := currentWorkers + 1
	log.Printf("Increasing the number of workers to %d", expectedWorkers)
	err = addWorker(os.Getenv("CLUSTER_NAME"))
	if err != nil {
		t.Fatalf("unexpected error setting number of workers to %d, %v", expectedWorkers, err)
	}

	if err := f.WaitForNodesUp(currentMasters + expectedWorkers); err != nil {
		t.Fatalf("unexpected error waiting for %d nodes up, %v", expectedWorkers, err)
	}
	log.Printf("%d worker nodes ready", expectedWorkers)

	// decrease number of workers
	expectedWorkers--
	log.Printf("Decreasing the number of workers to %d", expectedWorkers)
	err = removeWorker(os.Getenv("CLUSTER_NAME"))
	if err != nil {
		t.Fatalf("unexpected error setting number of workers to %d, %v", expectedWorkers, err)
	}

	if err := f.WaitForNodesUp(currentMasters + expectedWorkers); err != nil {
		t.Fatalf("unexpected error waiting for %d nodes up, %v", expectedWorkers, err)
	}
	log.Printf("%d worker nodes ready", expectedWorkers)
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

	logEntry := "creating AWS cloudformation stack: created"

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
	d := []byte(awsResourceChartValuesEnv)

	err := ioutil.WriteFile(awsResourceValuesFile, d, 0644)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func numberOfWorkers(clusterName string) (int, error) {
	cluster, err := f.AWSCluster(clusterName)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	return key.WorkerCount(*cluster), nil
}

func numberOfMasters(clusterName string) (int, error) {
	cluster, err := f.AWSCluster(clusterName)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	return key.MasterCount(*cluster), nil
}

func addWorker(clusterName string) error {
	cluster, err := f.AWSCluster(clusterName)
	if err != nil {
		return microerror.Mask(err)
	}

	newWorker := cluster.Spec.AWS.Workers[0]

	patch := make([]PatchSpec, 1)
	patch[0].Op = "add"
	patch[0].Path = "/spec/aws/workers/-"
	patch[0].Value = newWorker

	return f.ApplyAWSConfigPatch(patch, clusterName)
}

func removeWorker(clusterName string) error {
	patch := make([]PatchSpec, 1)
	patch[0].Op = "remove"
	patch[0].Path = "/spec/aws/workers/1"

	return f.ApplyAWSConfigPatch(patch, clusterName)
}

func createHostPeerVPC() error {
	log.Printf("Creating Host Peer VPC stack")

	hostVPCStack := `AWSTemplateFormatVersion: 2010-09-09
Description: CI Host Stack with Peering VPC and route tables
Resources:
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.11.0.0/16
      Tags:
      - Key: Name
        Value: ${CLUSTER_NAME}
  PeerRouteTable0:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: ${AWS_ROUTE_TABLE_0}
  PeerRouteTable1:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: ${AWS_ROUTE_TABLE_1}
Outputs:
  VPCID:
    Description: Accepter VPC ID
    Value: !Ref VPC

`
	os.Setenv("AWS_ROUTE_TABLE_0", os.Getenv("CLUSTER_NAME")+"_0")
	os.Setenv("AWS_ROUTE_TABLE_1", os.Getenv("CLUSTER_NAME")+"_1")
	stackName := "host-peer-" + os.Getenv("CLUSTER_NAME")
	stackInput := &cloudformation.CreateStackInput{
		StackName:        aws.String(stackName),
		TemplateBody:     aws.String(os.ExpandEnv(hostVPCStack)),
		TimeoutInMinutes: aws.Int64(2),
	}
	_, err := c.CF.CreateStack(stackInput)
	if err != nil {
		return microerror.Mask(err)
	}
	err = c.CF.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return microerror.Mask(err)
	}
	describeOutput, err := c.CF.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return microerror.Mask(err)
	}
	for _, o := range describeOutput.Stacks[0].Outputs {
		if *o.OutputKey == "VPCID" {
			os.Setenv("AWS_VPC_PEER_ID", *o.OutputValue)
			break
		}
	}
	log.Printf("Host Peer VPC stack created")
	return nil
}

func deleteHostPeerVPC() error {
	log.Printf("Deleting Host Peer VPC stack")

	_, err := c.CF.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String("host-peer-" + os.Getenv("CLUSTER_NAME")),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
