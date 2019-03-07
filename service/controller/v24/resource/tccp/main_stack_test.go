package tccp

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/v24/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v24/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v24/key"
)

func testConfig() Config {
	c := Config{}

	c.EncrypterBackend = "kms"
	c.G8sClient = fake.NewSimpleClientset()
	c.GuestPrivateSubnetMaskBits = 25
	c.GuestPublicSubnetMaskBits = 25
	c.HostClients = &adapter.Clients{}
	c.InstallationName = "myinstallation"
	c.Logger = microloggertest.New()

	return c
}

func statusWithAllocatedSubnet(cidr string, azs []string) v1alpha1.AWSConfigStatus {
	var statusAZs []v1alpha1.AWSConfigStatusAWSAvailabilityZone
	for _, az := range azs {
		statusAZs = append(statusAZs, v1alpha1.AWSConfigStatusAWSAvailabilityZone{
			Name: az,
		})
	}

	status := v1alpha1.AWSConfigStatus{
		AWS: v1alpha1.AWSConfigStatusAWS{
			AvailabilityZones: statusAZs,
		},
		Cluster: v1alpha1.StatusCluster{
			Network: v1alpha1.StatusClusterNetwork{
				CIDR: cidr,
			},
		},
	}

	return status
}

func TestMainGuestTemplateGetEmptyBody(t *testing.T) {
	t.Parallel()
	customObject := v1alpha1.AWSConfig{}

	c := testConfig()
	c.HostClients = &adapter.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
		STS: &adapter.STSClientMock{},
	}
	newResource, err := New(c)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	awsClients := aws.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
		KMS: &adapter.KMSClientMock{},
		STS: &adapter.STSClientMock{},
	}

	ctx := context.TODO()
	ctx = controllercontext.NewContext(ctx, controllercontext.Context{AWSClient: awsClients})

	_, err = newResource.getMainGuestTemplateBody(ctx, customObject, StackState{})
	if err == nil {
		t.Fatal("expected", nil, "got", err)
	}
}

func TestMainGuestTemplateExistingFields(t *testing.T) {
	t.Parallel()
	// customObject with example fields for both asg and launch config
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID:      "test-cluster",
				Version: "myversion",
				Kubernetes: v1alpha1.ClusterKubernetes{
					API: v1alpha1.ClusterKubernetesAPI{
						Domain:     "api.domain",
						SecurePort: 443,
					},
					IngressController: v1alpha1.ClusterKubernetesIngressController{
						Domain:       "ingress.domain",
						InsecurePort: 30010,
						SecurePort:   30011,
					},
				},
				Etcd: v1alpha1.ClusterEtcd{
					Domain: "etcd.domain",
				},
				Scaling: v1alpha1.ClusterScaling{
					Max: 3,
					Min: 3,
				},
			},
			AWS: v1alpha1.AWSConfigSpecAWS{
				API: v1alpha1.AWSConfigSpecAWSAPI{
					ELB: v1alpha1.AWSConfigSpecAWSAPIELB{
						IdleTimeoutSeconds: 3600,
					},
				},
				Region:            "eu-central-1",
				AZ:                "eu-central-1a",
				AvailabilityZones: 2,
				Masters: []v1alpha1.AWSConfigSpecAWSNode{
					{
						ImageID:      "ami-1234-master",
						InstanceType: "m3.large",
					},
				},
				Workers: []v1alpha1.AWSConfigSpecAWSNode{
					{
						DockerVolumeSizeGB: 150,
						ImageID:            "ami-1234-worker",
						InstanceType:       "m3.large",
					},
				},
				Ingress: v1alpha1.AWSConfigSpecAWSIngress{
					ELB: v1alpha1.AWSConfigSpecAWSIngressELB{
						IdleTimeoutSeconds: 60,
					},
				},
			},
		},
		Status: v1alpha1.AWSConfigStatus{
			AWS: v1alpha1.AWSConfigStatusAWS{
				AvailabilityZones: []v1alpha1.AWSConfigStatusAWSAvailabilityZone{
					v1alpha1.AWSConfigStatusAWSAvailabilityZone{
						Name: "eu-central-1a",
						Subnet: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
							Private: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
								CIDR: "10.1.1.0/26",
							},
							Public: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
								CIDR: "10.1.1.64/26",
							},
						},
					},
					v1alpha1.AWSConfigStatusAWSAvailabilityZone{
						Name: "eu-central-1b",
						Subnet: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
							Private: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
								CIDR: "10.1.1.128/26",
							},
							Public: v1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
								CIDR: "10.1.1.196/26",
							},
						},
					},
				},
			},
			Cluster: v1alpha1.StatusCluster{
				Network: v1alpha1.StatusClusterNetwork{
					CIDR: "10.1.1.0/24",
				},
			},
		},
	}

	imageID, err := key.ImageID(customObject)
	if err != nil {
		t.Fatalf("expected %#v got %#v", nil, err)
	}

	stackState := StackState{
		Name: key.MainGuestStackName(customObject),

		DockerVolumeResourceName:   key.DockerVolumeResourceName(customObject),
		MasterImageID:              imageID,
		MasterInstanceResourceName: key.MasterInstanceResourceName(customObject),
		MasterInstanceType:         key.MasterInstanceType(customObject),
		MasterCloudConfigVersion:   key.CloudConfigVersion,
		MasterInstanceMonitoring:   false,

		WorkerCloudConfigVersion: key.CloudConfigVersion,
		WorkerDockerVolumeSizeGB: key.WorkerDockerVolumeSizeGB(customObject),
		WorkerImageID:            imageID,
		WorkerInstanceMonitoring: true,
		WorkerInstanceType:       key.WorkerInstanceType(customObject),

		VersionBundleVersion: key.VersionBundleVersion(customObject),
	}

	cfg := testConfig()
	cfg.HostClients = &adapter.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
		STS: &adapter.STSClientMock{},
	}
	cfg.AdvancedMonitoringEC2 = true
	cfg.Route53Enabled = true
	newResource, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	awsClients := aws.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
		KMS: &adapter.KMSClientMock{},
		ELB: &adapter.ELBClientMock{},
		STS: &adapter.STSClientMock{},
	}

	ctx := context.TODO()
	ctx = controllercontext.NewContext(ctx, controllercontext.Context{AWSClient: awsClients})

	body, err := newResource.getMainGuestTemplateBody(ctx, customObject, stackState)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if !strings.Contains(body, "Description: Main Guest CloudFormation stack.") {
		t.Fatal("stack header not found")
	}

	if !strings.Contains(body, "  workerLaunchConfiguration:") {
		t.Fatal("launch configuration header not found")
	}

	if !strings.Contains(body, "  workerAutoScalingGroup:") {
		t.Fatal("asg header not found")
	}

	if !strings.Contains(body, "Key: k8s.io/cluster-autoscaler/enabled") {
		t.Fatal("cluster-autoscaler's tag missing from worker asg")
	}

	if !strings.Contains(body, fmt.Sprintf("Key: k8s.io/cluster-autoscaler/%s", key.ClusterID(customObject))) {
		t.Fatal("cluster-autoscaler's cluster tag missing from worker asg")
	}

	if !strings.Contains(body, "InstanceType: m3.large") {
		t.Fatal("launch configuration element not found")
	}

	if !strings.Contains(body, "Outputs:") {
		fmt.Println(body)
		t.Fatal("outputs header not found")
	}

	if !strings.Contains(body, key.MasterImageIDKey+":") {
		fmt.Println(body)
		t.Fatal("MasterImageID output element not found")
	}
	if !strings.Contains(body, key.MasterInstanceMonitoring+": true") {
		fmt.Println(body)
		t.Fatal("MasterInstanceMonitoring output element not found")
	}
	if !strings.Contains(body, key.MasterInstanceTypeKey+":") {
		fmt.Println(body)
		t.Fatal("MasterInstanceType output element not found")
	}
	if !strings.Contains(body, key.MasterCloudConfigVersionKey+":") {
		fmt.Println(body)
		t.Fatal("master CloudConfig version output element not found")
	}

	if !strings.Contains(body, key.WorkerImageIDKey+":") {
		fmt.Println(body)
		t.Fatal("WorkerImageID output element not found")
	}
	if !strings.Contains(body, key.WorkerInstanceTypeKey+":") {
		fmt.Println(body)
		t.Fatal("WorkerInstanceType output element not found")
	}
	if !strings.Contains(body, key.WorkerCloudConfigVersionKey+":") {
		fmt.Println(body)
		t.Fatal("worker CloudConfig version output element not found")
	}
	if !strings.Contains(body, key.WorkerInstanceMonitoring+": true") {
		fmt.Println(body)
		t.Fatal("WorkerInstanceMonitoring output element not found")
	}

	if !strings.Contains(body, "Value: "+key.CloudConfigVersion) {
		fmt.Println(body)
		t.Fatal("output element not found")
	}

	if !strings.Contains(body, workerRoleKey+":") {
		fmt.Println(body)
		t.Fatal("workerRole output element not found")
	}
	if !strings.Contains(body, "PolicyName: test-cluster-master") {
		fmt.Println(body)
		t.Fatal("PolicyName output element not found")
	}
	if !strings.Contains(body, "PolicyName: test-cluster-worker") {
		fmt.Println(body)
		t.Fatal("PolicyName output element not found")
	}
	if !strings.Contains(body, "ApiRecordSet:") {
		fmt.Println(body)
		t.Fatal("ApiRecordSet element not found")
	}
	if !strings.Contains(body, "EtcdRecordSet:") {
		fmt.Println(body)
		t.Fatal("EtcdRecordSet element not found")
	}
	if !strings.Contains(body, "IngressRecordSet:") {
		fmt.Println(body)
		t.Fatal("IngressRecordSet element not found")
	}
	if !strings.Contains(body, "IngressWildcardRecordSet:") {
		fmt.Println(body)
		t.Fatal("ingressWildcardRecordSet element not found")
	}
	if !strings.Contains(body, "MasterInstance") {
		fmt.Println(body)
		t.Fatal("MasterInstance element not found")
	}
	if !strings.Contains(body, "ApiLoadBalancer:") {
		fmt.Println(body)
		t.Fatal("ApiLoadBalancer element not found")
	}
	if !strings.Contains(body, "IngressLoadBalancer:") {
		fmt.Println(body)
		t.Fatal("IngressLoadBalancer element not found")
	}
	if !strings.Contains(body, "InternetGateway:") {
		fmt.Println(body)
		t.Fatal("InternetGateway element not found")
	}
	if !strings.Contains(body, "NATGateway:") {
		fmt.Println(body)
		t.Fatal("NATGateway element not found")
	}
	if !strings.Contains(body, "NATGateway01:") {
		fmt.Println(body)
		t.Fatal("NATGateway01 element not found")
	}
	if !strings.Contains(body, "PublicRouteTable:") {
		fmt.Println(body)
		t.Fatal("PublicRouteTable element not found")
	}
	if !strings.Contains(body, "PublicSubnet:") {
		fmt.Println(body)
		t.Fatal("PublicSubnet element not found")
	}
	if !strings.Contains(body, "PublicSubnet01:") {
		fmt.Println(body)
		t.Fatal("PublicSubnet01 element not found")
	}
	if !strings.Contains(body, "PrivateRouteTable:") {
		fmt.Println(body)
		t.Fatal("PrivateRouteTable element not found")
	}
	if !strings.Contains(body, "PrivateRouteTable01:") {
		fmt.Println(body)
		t.Fatal("PrivateRouteTable element not found")
	}
	if !strings.Contains(body, "PrivateSubnet:") {
		fmt.Println(body)
		t.Fatal("PrivateSubnet element not found")
	}
	if !strings.Contains(body, "PrivateSubnet01:") {
		fmt.Println(body)
		t.Fatal("PrivateSubnet01 element not found")
	}
	if !strings.Contains(body, "MasterSecurityGroup:") {
		fmt.Println(body)
		t.Fatal("MasterSecurityGroup element not found")
	}
	if !strings.Contains(body, "WorkerSecurityGroup:") {
		fmt.Println(body)
		t.Fatal("WorkerSecurityGroup element not found")
	}
	if !strings.Contains(body, "IngressSecurityGroup:") {
		fmt.Println(body)
		t.Fatal("IngressSecurityGroup element not found")
	}
	if !strings.Contains(body, " VPC:") {
		fmt.Println(body)
		t.Fatal("VPC element not found")
	}
	if !strings.Contains(body, "CidrBlock: 10.1.1.0/24") {
		fmt.Println(body)
		t.Fatal("CidrBlock element for VPC CIDR not found")
	}
	if !strings.Contains(body, "CidrBlock: 10.1.1.0/26") {
		fmt.Println(body)
		t.Fatal("CidrBlock element for private subnet 0 not found")
	}
	if !strings.Contains(body, "CidrBlock: 10.1.1.128/26") {
		fmt.Println(body)
		t.Fatal("CidrBlock element for private subnet 1 not found")
	}
	if !strings.Contains(body, "CidrBlock: 10.1.1.64/26") {
		fmt.Println(body)
		t.Fatal("CidrBlock element for public subnet 0 not found")
	}
	if !strings.Contains(body, "CidrBlock: 10.1.1.196/26") {
		fmt.Println(body)
		t.Fatal("CidrBlock element for public subnet 1 not found")
	}

	// arn depends on region
	if !strings.Contains(body, `Resource: "arn:aws:s3:::`) {
		fmt.Println(body)
		t.Fatal("ARN region dependent element not found")
	}

	// image ids should be fixed despite the values in the custom object
	if !strings.Contains(body, "ImageId: ami-015e6cb33a709348e") {
		fmt.Println(body)
		t.Fatal("Fixed image ID not found")
	}
}

func TestMainGuestTemplateRoute53Disabled(t *testing.T) {
	t.Parallel()
	// customObject with example fields for both asg and launch config
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID:      "test-cluster",
				Version: "myversion",
				Kubernetes: v1alpha1.ClusterKubernetes{
					API: v1alpha1.ClusterKubernetesAPI{
						Domain:     "api.domain",
						SecurePort: 443,
					},
					IngressController: v1alpha1.ClusterKubernetesIngressController{
						Domain:       "ingress.domain",
						InsecurePort: 30010,
						SecurePort:   30011,
					},
				},
				Etcd: v1alpha1.ClusterEtcd{
					Domain: "etcd.domain",
				},
				Scaling: v1alpha1.ClusterScaling{
					Max: 1,
					Min: 1,
				},
			},
			AWS: v1alpha1.AWSConfigSpecAWS{
				API: v1alpha1.AWSConfigSpecAWSAPI{
					ELB: v1alpha1.AWSConfigSpecAWSAPIELB{
						IdleTimeoutSeconds: 3600,
					},
				},
				Region: "eu-central-1",
				AZ:     "eu-central-1a",
				Masters: []v1alpha1.AWSConfigSpecAWSNode{
					{
						ImageID:      "ami-1234-master",
						InstanceType: "m3.large",
					},
				},
				Workers: []v1alpha1.AWSConfigSpecAWSNode{
					{
						DockerVolumeSizeGB: 150,
						ImageID:            "ami-1234-worker",
						InstanceType:       "m3.large",
					},
				},
				Ingress: v1alpha1.AWSConfigSpecAWSIngress{
					ELB: v1alpha1.AWSConfigSpecAWSIngressELB{
						IdleTimeoutSeconds: 60,
					},
				},
			},
		},
		Status: statusWithAllocatedSubnet("10.1.1.0/24", []string{"eu-central-1a"}),
	}

	imageID, err := key.ImageID(customObject)
	if err != nil {
		t.Fatalf("expected %#v got %#v", nil, err)
	}

	stackState := StackState{
		Name: key.MainGuestStackName(customObject),

		DockerVolumeResourceName:   key.DockerVolumeResourceName(customObject),
		MasterImageID:              imageID,
		MasterInstanceResourceName: key.MasterInstanceResourceName(customObject),
		MasterInstanceType:         key.MasterInstanceType(customObject),
		MasterCloudConfigVersion:   key.CloudConfigVersion,
		MasterInstanceMonitoring:   false,

		WorkerCloudConfigVersion: key.CloudConfigVersion,
		WorkerDockerVolumeSizeGB: key.WorkerDockerVolumeSizeGB(customObject),
		WorkerImageID:            imageID,
		WorkerInstanceMonitoring: true,
		WorkerInstanceType:       key.WorkerInstanceType(customObject),

		VersionBundleVersion: key.VersionBundleVersion(customObject),
	}

	cfg := testConfig()
	cfg.HostClients = &adapter.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
		STS: &adapter.STSClientMock{},
	}
	cfg.Route53Enabled = false
	newResource, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	awsClients := aws.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
		KMS: &adapter.KMSClientMock{},
		ELB: &adapter.ELBClientMock{},
		STS: &adapter.STSClientMock{},
	}

	ctx := context.TODO()
	ctx = controllercontext.NewContext(ctx, controllercontext.Context{AWSClient: awsClients})

	body, err := newResource.getMainGuestTemplateBody(ctx, customObject, stackState)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if strings.Contains(body, "ApiRecordSet:") {
		fmt.Println(body)
		t.Fatal("ApiRecordSet element found")
	}
	if strings.Contains(body, "EtcdRecordSet:") {
		fmt.Println(body)
		t.Fatal("EtcdRecordSet element found")
	}
	if strings.Contains(body, "IngressRecordSet:") {
		fmt.Println(body)
		t.Fatal("IngressRecordSet element found")
	}
	if strings.Contains(body, "IngressWildcardRecordSet:") {
		fmt.Println(body)
		t.Fatal("ingressWildcardRecordSet element found")
	}
}

func TestMainGuestTemplateChinaRegion(t *testing.T) {
	t.Parallel()
	// customObject with example fields for both asg and launch config
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID:      "test-cluster",
				Version: "myversion",
				Kubernetes: v1alpha1.ClusterKubernetes{
					API: v1alpha1.ClusterKubernetesAPI{
						Domain:     "api.domain",
						SecurePort: 443,
					},
					IngressController: v1alpha1.ClusterKubernetesIngressController{
						Domain:       "ingress.domain",
						InsecurePort: 30010,
						SecurePort:   30011,
					},
				},
				Etcd: v1alpha1.ClusterEtcd{
					Domain: "etcd.domain",
				},
				Scaling: v1alpha1.ClusterScaling{
					Max: 1,
					Min: 1,
				},
			},
			AWS: v1alpha1.AWSConfigSpecAWS{
				API: v1alpha1.AWSConfigSpecAWSAPI{
					ELB: v1alpha1.AWSConfigSpecAWSAPIELB{
						IdleTimeoutSeconds: 3600,
					},
				},
				Region: "cn-north-1",
				AZ:     "cn-north-1a",
				Masters: []v1alpha1.AWSConfigSpecAWSNode{
					{
						ImageID:      "ami-1234-master",
						InstanceType: "m3.large",
					},
				},
				Workers: []v1alpha1.AWSConfigSpecAWSNode{
					{
						ImageID:      "ami-1234-worker",
						InstanceType: "m3.large",
					},
				},
				Ingress: v1alpha1.AWSConfigSpecAWSIngress{
					ELB: v1alpha1.AWSConfigSpecAWSIngressELB{
						IdleTimeoutSeconds: 60,
					},
				},
			},
		},
		Status: statusWithAllocatedSubnet("10.1.1.0/24", []string{"cn-north-1a"}),
	}

	imageID, err := key.ImageID(customObject)
	if err != nil {
		t.Fatalf("expected %#v got %#v", nil, err)
	}

	stackState := StackState{
		Name: key.MainGuestStackName(customObject),

		DockerVolumeResourceName:   key.DockerVolumeResourceName(customObject),
		MasterImageID:              imageID,
		MasterInstanceResourceName: key.MasterInstanceResourceName(customObject),
		MasterInstanceType:         key.MasterInstanceType(customObject),
		MasterCloudConfigVersion:   key.CloudConfigVersion,
		MasterInstanceMonitoring:   false,

		WorkerCloudConfigVersion: key.CloudConfigVersion,
		WorkerImageID:            imageID,
		WorkerInstanceMonitoring: true,
		WorkerInstanceType:       key.WorkerInstanceType(customObject),

		VersionBundleVersion: key.VersionBundleVersion(customObject),
	}

	cfg := testConfig()
	cfg.HostClients = &adapter.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
		STS: &adapter.STSClientMock{},
	}
	cfg.Route53Enabled = false
	newResource, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	awsClients := aws.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
		KMS: &adapter.KMSClientMock{},
		ELB: &adapter.ELBClientMock{},
		STS: &adapter.STSClientMock{},
	}

	ctx := context.TODO()
	ctx = controllercontext.NewContext(ctx, controllercontext.Context{AWSClient: awsClients})

	body, err := newResource.getMainGuestTemplateBody(ctx, customObject, stackState)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	// arn depends on region
	if !strings.Contains(body, `Resource: "arn:aws-cn:s3:::`) {
		fmt.Println(body)
		t.Fatal("ARN region dependent element not found")
	}
}
