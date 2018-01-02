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

func TestMainTemplateGetEmptyBody(t *testing.T) {
	customObject := v1alpha1.AWSConfig{}
	cfg := testConfig()

	newResource, err := New(cfg)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	_, err = newResource.getMainTemplateBody(customObject)
	if err == nil {
		t.Error("error didn't happen")
	}
}

func TestMainTemplateExistingFields(t *testing.T) {
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
						Domain: "ingress.domain",
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
	newResource, err := New(cfg)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	body, err := newResource.getMainTemplateBody(customObject)

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if !strings.Contains(body, "Description: Main CloudFormation stack.") {
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
	if !strings.Contains(body, "InternetGateway:") {
		fmt.Println(body)
		t.Error("InternetGateway element not found")
	}
	if !strings.Contains(body, "NATGateway:") {
		fmt.Println(body)
		t.Error("NATGateway element not found")
	}
	if !strings.Contains(body, "SubnetId: subnet-1234") {
		fmt.Println(body)
		t.Error("NATGateway subnet id property not found")
	}
}
