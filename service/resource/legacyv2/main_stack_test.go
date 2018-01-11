package legacyv2

import (
	"fmt"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeytpr"
	"k8s.io/client-go/kubernetes/fake"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/cloudconfigv2"
	"github.com/giantswarm/aws-operator/service/resource/legacyv2/adapter"
)

func testConfig() Config {
	resourceConfig := DefaultConfig()
	resourceConfig.Clients = &adapter.Clients{}
	resourceConfig.HostClients = &adapter.Clients{}
	resourceConfig.Logger = microloggertest.New()
	resourceConfig.CloudConfig = &cloudconfigv2.CloudConfig{}
	resourceConfig.CertWatcher = &certificatetpr.Service{}
	resourceConfig.KeyWatcher = &randomkeytpr.Service{}
	resourceConfig.K8sClient = fake.NewSimpleClientset()
	resourceConfig.InstallationName = "myinstallation"
	resourceConfig.AwsConfig = awsutil.Config{AccessKeyID: "myaccessKey"}
	resourceConfig.AwsHostConfig = awsutil.Config{AccessKeyID: "myaccessKey"}
	resourceConfig.PubKeyFile = "mypubkeyfile"
	return resourceConfig
}

func TestMainGuestTemplateGetEmptyBody(t *testing.T) {
	customObject := v1alpha1.AWSConfig{}
	cfg := testConfig()
	cfg.Clients = &adapter.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
		KMS: &adapter.KMSClientMock{},
	}
	cfg.HostClients = &adapter.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
	}
	newResource, err := New(cfg)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	_, err = newResource.getMainGuestTemplateBody(customObject)
	if err == nil {
		t.Error("error didn't happen")
	}
}

func TestMainGuestTemplateExistingFields(t *testing.T) {
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
			},
			AWS: v1alpha1.AWSConfigSpecAWS{
				API: v1alpha1.AWSConfigSpecAWSAPI{
					ELB: v1alpha1.AWSConfigSpecAWSAPIELB{
						IdleTimeoutSeconds: 3600,
					},
				},
				AZ: "myaz",
				Masters: []v1alpha1.AWSConfigSpecAWSNode{
					v1alpha1.AWSConfigSpecAWSNode{
						ImageID: "myimageid",
					},
				},
				Workers: []v1alpha1.AWSConfigSpecAWSNode{
					v1alpha1.AWSConfigSpecAWSNode{
						ImageID: "myimageid",
					},
				},
				Ingress: v1alpha1.AWSConfigSpecAWSIngress{
					ELB: v1alpha1.AWSConfigSpecAWSIngressELB{
						IdleTimeoutSeconds: 60,
					},
				},
				VPC: v1alpha1.AWSConfigSpecAWSVPC{
					CIDR:              "10.1.1.0/24",
					PublicSubnetCIDR:  "10.1.1.0/25",
					PrivateSubnetCIDR: "10.1.2.0/25",
				},
			},
		},
	}

	cfg := testConfig()
	cfg.Clients = &adapter.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
		KMS: &adapter.KMSClientMock{},
		ELB: &adapter.ELBClientMock{},
	}
	cfg.HostClients = &adapter.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
	}
	newResource, err := New(cfg)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	body, err := newResource.getMainGuestTemplateBody(customObject)

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if !strings.Contains(body, "Description: Main Guest CloudFormation stack.") {
		t.Error("stack header not found")
	}

	if !strings.Contains(body, "  workerLaunchConfiguration:") {
		t.Error("launch configuration header not found")
	}

	if !strings.Contains(body, "  workerAutoScalingGroup:") {
		t.Error("asg header not found")
	}

	if !strings.Contains(body, "ImageId: myimageid") {
		t.Error("launch configuration element not found")
	}

	if !strings.Contains(body, "AvailabilityZones: [myaz]") {
		fmt.Println(body)
		t.Error("asg element not found")
	}

	if !strings.Contains(body, "Outputs:") {
		fmt.Println(body)
		t.Error("outputs header not found")
	}

	if !strings.Contains(body, workersOutputKey+":") {
		fmt.Println(body)
		t.Error("workers output element not found")
	}
	if !strings.Contains(body, imageIDOutputKey+":") {
		fmt.Println(body)
		t.Error("imageID output element not found")
	}
	if !strings.Contains(body, clusterVersionOutputKey+":") {
		fmt.Println(body)
		t.Error("clusterVersion output element not found")
	}
	if !strings.Contains(body, "Value: myversion") {
		fmt.Println(body)
		t.Error("output element not found")
	}
	if !strings.Contains(body, workerRoleKey+":") {
		fmt.Println(body)
		t.Error("workerRole output element not found")
	}
	if !strings.Contains(body, "PolicyName: test-cluster-master") {
		fmt.Println(body)
		t.Error("PolicyName output element not found")
	}
	if !strings.Contains(body, "PolicyName: test-cluster-worker") {
		fmt.Println(body)
		t.Error("PolicyName output element not found")
	}
	if !strings.Contains(body, "ApiRecordSet:") {
		fmt.Println(body)
		t.Error("ApiRecordSet element not found")
	}
	if !strings.Contains(body, "EtcdRecordSet:") {
		fmt.Println(body)
		t.Error("EtcdRecordSet element not found")
	}
	if !strings.Contains(body, "IngressRecordSet:") {
		fmt.Println(body)
		t.Error("IngressRecordSet element not found")
	}
	if !strings.Contains(body, "IngressWildcardRecordSet:") {
		fmt.Println(body)
		t.Error("ingressWildcardRecordSet element not found")
	}
	if !strings.Contains(body, "MasterInstance:") {
		fmt.Println(body)
		t.Error("MasterInstance element not found")
	}
	if !strings.Contains(body, "ApiLoadBalancer:") {
		fmt.Println(body)
		t.Error("ApiLoadBalancer element not found")
	}
	if !strings.Contains(body, "IngressLoadBalancer:") {
		fmt.Println(body)
		t.Error("IngressLoadBalancer element not found")
	}
	if !strings.Contains(body, "InternetGateway:") {
		fmt.Println(body)
		t.Error("InternetGateway element not found")
	}
	if !strings.Contains(body, "NATGateway:") {
		fmt.Println(body)
		t.Error("NATGateway element not found")
	}
	if !strings.Contains(body, "PublicRouteTable:") {
		fmt.Println(body)
		t.Error("PublicRouteTable element not found")
	}
	if !strings.Contains(body, "PublicSubnet:") {
		fmt.Println(body)
		t.Error("PublicSubnet element not found")
	}
	if !strings.Contains(body, "PrivateRouteTable:") {
		fmt.Println(body)
		t.Error("PrivateRouteTable element not found")
	}
	if !strings.Contains(body, "PrivateSubnet:") {
		fmt.Println(body)
		t.Error("PrivateSubnet element not found")
	}
	if !strings.Contains(body, "MasterSecurityGroup:") {
		fmt.Println(body)
		t.Error("MasterSecurityGroup element not found")
	}
	if !strings.Contains(body, "WorkerSecurityGroup:") {
		fmt.Println(body)
		t.Error("WorkerSecurityGroup element not found")
	}
	if !strings.Contains(body, "IngressSecurityGroup:") {
		fmt.Println(body)
		t.Error("IngressSecurityGroup element not found")
	}
	if !strings.Contains(body, " VPC:") {
		fmt.Println(body)
		t.Error("VPC element not found")
	}
	if !strings.Contains(body, "CidrBlock: 10.1.1.0/24") {
		fmt.Println(body)
		t.Error("CidrBlock element not found")
	}
}

func TestMainHostPreTemplateExistingFields(t *testing.T) {
	// customObject with example fields for both asg and launch config
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	cfg := testConfig()
	cfg.Clients = &adapter.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
	}
	cfg.HostClients = &adapter.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
	}
	newResource, err := New(cfg)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	body, err := newResource.getMainHostPreTemplateBody(customObject)

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if !strings.Contains(body, "Description: Main Host Pre-Guest CloudFormation stack.") {
		fmt.Println(body)
		t.Error("stack header not found")
	}

	if !strings.Contains(body, "  PeerRole:") {
		fmt.Println(body)
		t.Error("peer role header not found")
	}

	if !strings.Contains(body, "  RoleName: test-cluster-vpc-peer-access") {
		fmt.Println(body)
		t.Error("role name not found")
	}
}

func TestMainHostPostTemplateExistingFields(t *testing.T) {
	// customObject with example fields for both asg and launch config
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
			AWS: v1alpha1.AWSConfigSpecAWS{
				VPC: v1alpha1.AWSConfigSpecAWSVPC{
					RouteTableNames: []string{
						"route_table_1",
						"route_table_2",
					},
					PeerID: "mypeerid",
				},
			},
		},
	}

	cfg := testConfig()
	cfg.Clients = &adapter.Clients{
		EC2: &adapter.EC2ClientMock{},
	}
	cfg.HostClients = &adapter.Clients{
		EC2: &adapter.EC2ClientMock{},
		IAM: &adapter.IAMClientMock{},
	}
	newResource, err := New(cfg)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	body, err := newResource.getMainHostPostTemplateBody(customObject)

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if !strings.Contains(body, "Description: Main Host Post-Guest CloudFormation stack.") {
		fmt.Println(body)
		t.Error("stack header not found")
	}

	if !strings.Contains(body, "  PrivateRoute1:") {
		fmt.Println(body)
		t.Error("route header not found")
	}
}
