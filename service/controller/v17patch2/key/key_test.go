package key

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/stretchr/testify/assert"
)

func Test_AutoScalingGroupName(t *testing.T) {
	t.Parallel()
	expectedName := "test-cluster-worker"
	groupName := "worker"

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
				Customer: v1alpha1.ClusterCustomer{
					ID: "test-customer",
				},
			},
		},
	}

	if AutoScalingGroupName(customObject, groupName) != expectedName {
		t.Fatalf("Expected auto scaling group name %s but was %s", expectedName, AutoScalingGroupName(customObject, groupName))
	}
}

func Test_AvailabilityZone(t *testing.T) {
	t.Parallel()
	expectedAZ := "eu-central-1a"

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			AWS: v1alpha1.AWSConfigSpecAWS{
				AZ: "eu-central-1a",
			},
		},
	}

	if AvailabilityZone(customObject) != expectedAZ {
		t.Fatalf("Expected availability zone %s but was %s", expectedAZ, AvailabilityZone(customObject))
	}
}

func Test_BaseDomain(t *testing.T) {
	t.Parallel()
	expectedBaseDomain := "installtion.eu-central-1.aws.gigantic.io"

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			AWS: v1alpha1.AWSConfigSpecAWS{
				HostedZones: v1alpha1.AWSConfigSpecAWSHostedZones{
					API: v1alpha1.AWSConfigSpecAWSHostedZonesZone{
						Name: "installtion.eu-central-1.aws.gigantic.io",
					},
				},
			},
		},
	}

	baseDomain := BaseDomain(customObject)
	if baseDomain != expectedBaseDomain {
		t.Fatalf("BaseDomain == %q, want %q", baseDomain, expectedBaseDomain)
	}
}

func Test_BucketName(t *testing.T) {
	t.Parallel()
	accountID := "1234567890"
	expectedName := "1234567890-g8s-test-cluster"

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	if BucketName(customObject, accountID) != expectedName {
		t.Fatalf("Expected bucket name %s but was %s", expectedName, BucketName(customObject, accountID))
	}
}

func Test_ClusterID(t *testing.T) {
	t.Parallel()
	expectedID := "test-cluster"

	cluster := v1alpha1.Cluster{
		ID: expectedID,
		Customer: v1alpha1.ClusterCustomer{
			ID: "test-customer",
		},
	}

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: cluster,
		},
	}

	if ClusterID(customObject) != expectedID {
		t.Fatalf("Expected cluster ID %s but was %s", expectedID, ClusterID(customObject))
	}
}

func Test_ClusterCustomer(t *testing.T) {
	t.Parallel()
	expectedCustomer := "test-customer"

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
				Customer: v1alpha1.ClusterCustomer{
					ID: "test-customer",
				},
			},
		},
	}

	if ClusterCustomer(customObject) != expectedCustomer {
		t.Fatalf("Expected customer ID %s but was %s", expectedCustomer, ClusterCustomer(customObject))
	}
}

func Test_ClusterCloudProviderTag(t *testing.T) {
	t.Parallel()
	expectedID := "test-cluster"
	expectedTag := "kubernetes.io/cluster/test-cluster"

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: expectedID,
			},
		},
	}

	if ClusterCloudProviderTag(customObject) != expectedTag {
		t.Fatalf("Expected cloud provider tag %s but was %s", expectedTag, ClusterCloudProviderTag(customObject))
	}
}

func Test_ClusterNamespace(t *testing.T) {
	t.Parallel()
	expectedID := "test-cluster"

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: expectedID,
			},
		},
	}

	if ClusterNamespace(customObject) != expectedID {
		t.Fatalf("Expected cluster ID %s but was %s", expectedID, ClusterNamespace(customObject))
	}
}

func Test_ClusterOrganization(t *testing.T) {
	t.Parallel()
	expectedOrg := "test-org"

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
				Customer: v1alpha1.ClusterCustomer{
					ID: "test-org",
				},
			},
		},
	}

	if ClusterOrganization(customObject) != expectedOrg {
		t.Fatalf("Expected organization ID %s but was %s", expectedOrg, ClusterOrganization(customObject))
	}
}

func Test_ClusterTags(t *testing.T) {
	t.Parallel()
	installName := "test-install"

	expectedID := "test-cluster"
	expectedTags := map[string]string{
		"kubernetes.io/cluster/test-cluster": "owned",
		"giantswarm.io/cluster":              "test-cluster",
		"giantswarm.io/installation":         "test-install",
		"giantswarm.io/organization":         "test-org",
	}

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: expectedID,
				// Organization uses Customer until its renamed in the CRD.
				Customer: v1alpha1.ClusterCustomer{
					ID: "test-org",
				},
			},
		},
	}

	if !reflect.DeepEqual(expectedTags, ClusterTags(customObject, installName)) {
		t.Fatalf("Expected cluster tags %v but was %v", expectedTags, ClusterTags(customObject, installName))
	}
}

func Test_ClusterVersion(t *testing.T) {
	t.Parallel()
	expectedVersion := "v_0_1_0"

	cluster := v1alpha1.Cluster{
		Version: expectedVersion,
	}

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: cluster,
		},
	}

	if ClusterVersion(customObject) != expectedVersion {
		t.Fatalf("Expected cluster version %s but was %s", expectedVersion, ClusterVersion(customObject))
	}
}

func Test_EC2ServiceDomain(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description              string
		customObject             v1alpha1.AWSConfig
		expectedEC2ServiceDomain string
	}{
		{
			description: "basic match",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "eu-central-1",
					},
				},
			},
			expectedEC2ServiceDomain: "ec2.amazonaws.com",
		},
		{
			description: "different region",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "us-west-2",
					},
				},
			},
			expectedEC2ServiceDomain: "ec2.amazonaws.com",
		},
		{
			description: "china region",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "cn-north-1",
					},
				},
			},
			expectedEC2ServiceDomain: "ec2.amazonaws.com.cn",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			ec2ServiceDomain := EC2ServiceDomain(tc.customObject)

			if tc.expectedEC2ServiceDomain != ec2ServiceDomain {
				t.Errorf("unexpected EC2 service domain, expecting %q, want %q", tc.expectedEC2ServiceDomain, ec2ServiceDomain)
			}
		})
	}
}

func Test_DockerVolumeResourceName_Format(t *testing.T) {
	t.Parallel()

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	n := DockerVolumeResourceName(customObject)

	prefix := "DockerVolume"

	if !strings.HasPrefix(n, prefix) {
		t.Fatalf("expected %s to have prefix %s", n, prefix)
	}
}

func Test_DockerVolumeResourceName_Inequivalence(t *testing.T) {
	t.Parallel()

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	n1 := DockerVolumeResourceName(customObject)
	time.Sleep(1 * time.Millisecond)
	n2 := DockerVolumeResourceName(customObject)

	if n1 == n2 {
		t.Fatalf("expected %s to differ from %s", n1, n2)
	}
}

func Test_EtcdVolumeName(t *testing.T) {
	t.Parallel()
	expectedName := "test-cluster-etcd"

	cluster := v1alpha1.Cluster{
		ID: "test-cluster",
		Customer: v1alpha1.ClusterCustomer{
			ID: "test-customer",
		},
	}

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: cluster,
		},
	}

	if EtcdVolumeName(customObject) != expectedName {
		t.Fatalf("Expected Etcd volume name %s but was %s", expectedName, EtcdVolumeName(customObject))
	}
}

func Test_IngressControllerInsecurePort(t *testing.T) {
	t.Parallel()
	expectedPort := 30010
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				Kubernetes: v1alpha1.ClusterKubernetes{
					IngressController: v1alpha1.ClusterKubernetesIngressController{
						InsecurePort: expectedPort,
					},
				},
			},
		},
	}

	if IngressControllerInsecurePort(customObject) != expectedPort {
		t.Fatalf("Expected ingress controller insecure port %d but was %d", expectedPort, IngressControllerInsecurePort(customObject))
	}
}

func Test_IngressControllerSecurePort(t *testing.T) {
	t.Parallel()
	expectedPort := 30011
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				Kubernetes: v1alpha1.ClusterKubernetes{
					IngressController: v1alpha1.ClusterKubernetesIngressController{
						SecurePort: expectedPort,
					},
				},
			},
		},
	}

	if IngressControllerSecurePort(customObject) != expectedPort {
		t.Fatalf("Expected ingress controller secure port %d but was %d", expectedPort, IngressControllerSecurePort(customObject))
	}
}

func Test_IsChinaRegion(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description    string
		customObject   v1alpha1.AWSConfig
		expectedResult bool
	}{
		{
			description: "non china region",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "eu-central-1",
					},
				},
			},
			expectedResult: false,
		},
		{
			description: "different non china region",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "us-west-2",
					},
				},
			},
			expectedResult: false,
		},
		{
			description: "china region",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "cn-north-1",
					},
				},
			},
			expectedResult: true,
		},
		{
			description: "different china region",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "cn-northwest-1",
					},
				},
			},
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if tc.expectedResult != IsChinaRegion(tc.customObject) {
				t.Errorf("unexpected result, expecting %t, want %t", tc.expectedResult, IsChinaRegion(tc.customObject))
			}
		})
	}
}

func Test_KubernetesAPISecurePort(t *testing.T) {
	t.Parallel()
	expectedPort := 443
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				Kubernetes: v1alpha1.ClusterKubernetes{
					API: v1alpha1.ClusterKubernetesAPI{
						SecurePort: expectedPort,
					},
				},
			},
		},
	}

	if KubernetesAPISecurePort(customObject) != expectedPort {
		t.Fatalf("Expected kubernetes api secure port %d but was %d", expectedPort, KubernetesAPISecurePort(customObject))
	}
}

func Test_MasterImageID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		customObject    v1alpha1.AWSConfig
		expectedImageID string
	}{
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							{
								ImageID:      "ami-d60ad6b9",
								InstanceType: "m3.medium",
							},
						},
					},
				},
			},
			expectedImageID: "ami-d60ad6b9",
		},
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{},
			},
			expectedImageID: "",
		},
	}

	for _, tc := range tests {
		if MasterImageID(tc.customObject) != tc.expectedImageID {
			t.Fatalf("Expected master image ID %s but was %s", tc.expectedImageID, MasterImageID(tc.customObject))
		}
	}
}

func Test_MasterInstanceName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		customObject         v1alpha1.AWSConfig
		expectedInstanceName string
	}{
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
				},
			},
			expectedInstanceName: "test-cluster-master",
		},
	}

	for _, tc := range tests {
		if MasterInstanceName(tc.customObject) != tc.expectedInstanceName {
			t.Fatalf("Expected master instance name %s but was %s", tc.expectedInstanceName, MasterInstanceName(tc.customObject))
		}
	}
}

func Test_MasterInstanceResourceName_Format(t *testing.T) {
	t.Parallel()

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	n := MasterInstanceResourceName(customObject)

	prefix := "MasterInstance"

	if !strings.HasPrefix(n, prefix) {
		t.Fatalf("expected %s to have prefix %s", n, prefix)
	}
}

func Test_MasterInstanceResourceName_Inequivalence(t *testing.T) {
	t.Parallel()

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	n1 := MasterInstanceResourceName(customObject)
	time.Sleep(1 * time.Millisecond)
	n2 := MasterInstanceResourceName(customObject)

	if n1 == n2 {
		t.Fatalf("expected %s to differ from %s", n1, n2)
	}
}

func Test_MasterInstanceType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		customObject         v1alpha1.AWSConfig
		expectedInstanceType string
	}{
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							{
								ImageID:      "ami-d60ad6b9",
								InstanceType: "m3.medium",
							},
						},
					},
				},
			},
			expectedInstanceType: "m3.medium",
		},
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							{},
						},
					},
				},
			},
			expectedInstanceType: "",
		},
	}

	for _, tc := range tests {
		if MasterInstanceType(tc.customObject) != tc.expectedInstanceType {
			t.Fatalf("Expected master instance type %s but was %s", tc.expectedInstanceType, MasterInstanceType(tc.customObject))
		}
	}
}

func Test_Region(t *testing.T) {
	t.Parallel()
	expectedRegion := "eu-central-1"

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			AWS: v1alpha1.AWSConfigSpecAWS{
				Region: "eu-central-1",
			},
		},
	}

	if Region(customObject) != expectedRegion {
		t.Fatalf("Expected region %s but was %s", expectedRegion, Region(customObject))
	}
}

func Test_RouteTableName(t *testing.T) {
	t.Parallel()
	expectedName := "test-cluster-private"
	suffix := "private"

	cluster := v1alpha1.Cluster{
		ID: "test-cluster",
	}

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: cluster,
		},
	}

	if RouteTableName(customObject, suffix) != expectedName {
		t.Fatalf("Expected route table name %s but was %s", expectedName, RouteTableName(customObject, suffix))
	}

}

func Test_S3ServiceDomain(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description      string
		customObject     v1alpha1.AWSConfig
		expectedS3Domain string
	}{
		{
			description: "basic match",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "eu-central-1",
					},
				},
			},
			expectedS3Domain: "s3.eu-central-1.amazonaws.com",
		},
		{
			description: "different region",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "us-west-2",
					},
				},
			},
			expectedS3Domain: "s3.us-west-2.amazonaws.com",
		},
		{
			description: "china region",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "cn-north-1",
					},
				},
			},
			expectedS3Domain: "s3.cn-north-1.amazonaws.com.cn",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			s3Domain := S3ServiceDomain(tc.customObject)

			if tc.expectedS3Domain != s3Domain {
				t.Errorf("unexpected S3 service domain, expecting %q, want %q", tc.expectedS3Domain, s3Domain)
			}
		})
	}
}

func Test_SecurityGroupName(t *testing.T) {
	t.Parallel()
	expectedName := "test-cluster-worker"
	groupName := "worker"

	cluster := v1alpha1.Cluster{
		ID: "test-cluster",
		Customer: v1alpha1.ClusterCustomer{
			ID: "test-customer",
		},
	}

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: cluster,
		},
	}

	if SecurityGroupName(customObject, groupName) != expectedName {
		t.Fatalf("Expected security group name %s but was %s", expectedName, SecurityGroupName(customObject, groupName))
	}
}

func Test_SubnetName(t *testing.T) {
	t.Parallel()
	expectedName := "test-cluster-private"
	suffix := "private"

	cluster := v1alpha1.Cluster{
		ID: "test-cluster",
	}

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: cluster,
		},
	}

	if SubnetName(customObject, suffix) != expectedName {
		t.Fatalf("Expected subnet name %s but was %s", expectedName, SubnetName(customObject, suffix))
	}

}

func Test_WorkerCount(t *testing.T) {
	t.Parallel()
	expectedCount := 2

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			AWS: v1alpha1.AWSConfigSpecAWS{
				Workers: []v1alpha1.AWSConfigSpecAWSNode{
					{
						InstanceType: "m3.medium",
					},
					{
						InstanceType: "m3.medium",
					},
				},
			},
		},
	}

	if WorkerCount(customObject) != expectedCount {
		t.Fatalf("Expected worker count %d but was %d", expectedCount, WorkerCount(customObject))
	}
}

func Test_WorkerDockerVolumeSizeGB(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		customObject v1alpha1.AWSConfig
		expectedSize int
	}{
		{
			name: "case 0: worker with 350GB docker volume",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							{
								DockerVolumeSizeGB: 350,
							},
						},
					},
				},
			},
			expectedSize: 350,
		},
		{
			name: "case 1: no workers",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Workers: []v1alpha1.AWSConfigSpecAWSNode{},
					},
				},
			},
			expectedSize: 100,
		},
		{
			name: "case 2: missing field for worker",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							{},
						},
					},
				},
			},
			expectedSize: 100,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sz := WorkerDockerVolumeSizeGB(tc.customObject)

			if sz != tc.expectedSize {
				t.Fatalf("Expected worker docker volume size  %d but was %d", tc.expectedSize, sz)
			}
		})
	}
}

func Test_WorkerImageID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		customObject    v1alpha1.AWSConfig
		expectedImageID string
	}{
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							{
								ImageID:      "ami-d60ad6b9",
								InstanceType: "m3.medium",
							},
						},
					},
				},
			},
			expectedImageID: "ami-d60ad6b9",
		},
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Workers: []v1alpha1.AWSConfigSpecAWSNode{},
					},
				},
			},
			expectedImageID: "",
		},
	}

	for _, tc := range tests {
		if WorkerImageID(tc.customObject) != tc.expectedImageID {
			t.Fatalf("Expected worker image ID %s but was %s", tc.expectedImageID, WorkerImageID(tc.customObject))
		}
	}
}

func Test_WorkerInstanceType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		customObject         v1alpha1.AWSConfig
		expectedInstanceType string
	}{
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							{
								ImageID:      "ami-d60ad6b9",
								InstanceType: "m3.medium",
							},
						},
					},
				},
			},
			expectedInstanceType: "m3.medium",
		},
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Workers: []v1alpha1.AWSConfigSpecAWSNode{},
					},
				},
			},
			expectedInstanceType: "",
		},
	}

	for _, tc := range tests {
		if WorkerInstanceType(tc.customObject) != tc.expectedInstanceType {
			t.Fatalf("Expected worker instance type %s but was %s", tc.expectedInstanceType, WorkerInstanceType(tc.customObject))
		}
	}
}

func Test_MainGuestStackName(t *testing.T) {
	t.Parallel()
	expected := "cluster-xyz-guest-main"

	cluster := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "xyz",
			},
		},
	}

	actual := MainGuestStackName(cluster)
	if actual != expected {
		t.Fatalf("Expected main stack name %s but was %s", expected, actual)
	}
}

func Test_MainHostPreStackName(t *testing.T) {
	t.Parallel()
	expected := "cluster-xyz-host-setup"

	cluster := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "xyz",
			},
		},
	}

	actual := MainHostPreStackName(cluster)
	if actual != expected {
		t.Fatalf("Expected main stack name %s but was %s", expected, actual)
	}
}

func Test_MainHostPostStackName(t *testing.T) {
	t.Parallel()
	expected := "cluster-xyz-host-main"

	cluster := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "xyz",
			},
		},
	}

	actual := MainHostPostStackName(cluster)
	if actual != expected {
		t.Fatalf("Expected main stack name %s but was %s", expected, actual)
	}
}

func Test_InstanceProfileName(t *testing.T) {
	t.Parallel()
	expectedName := "test-cluster-worker-EC2-K8S-Role"
	profileType := "worker"

	cluster := v1alpha1.Cluster{
		ID: "test-cluster",
	}

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: cluster,
		},
	}

	if InstanceProfileName(customObject, profileType) != expectedName {
		t.Fatalf("Expected instance profile name '%s' but was '%s'", expectedName, InstanceProfileName(customObject, profileType))
	}
}

func TestLoadBalancerName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc       string
		domainName string
		tpo        v1alpha1.AWSConfig
		res        string
		err        error
	}{
		{
			desc:       "works",
			domainName: "component.foo.bar.example.com",
			tpo: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "foo-customer",
					},
				},
			},
			res: "foo-customer-component",
		},
		{
			desc:       "also works",
			domainName: "component.of.a.well.formed.domain",
			tpo: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "quux-the-customer",
					},
				},
			},
			res: "quux-the-customer-component",
		},
		{
			desc:       "missing ID key in cloudconfig",
			domainName: "component.foo.bar.example.com",
			tpo: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "",
					},
				},
			},
			res: "",
			err: missingCloudConfigKeyError,
		},
		{
			desc:       "malformed domain name",
			domainName: "not a domain name",
			tpo: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "foo-customer",
					},
				},
			},
			res: "",
			err: malformedCloudConfigKeyError,
		},
		{
			desc:       "missing domain name",
			domainName: "",
			tpo: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "foo-customer",
					},
				},
			},
			res: "",
			err: malformedCloudConfigKeyError,
		},
	}

	for _, tc := range tests {
		res, err := LoadBalancerName(tc.domainName, tc.tpo)

		if err != nil {
			underlying := microerror.Cause(err)
			assert.Equal(t, tc.err, underlying, fmt.Sprintf("[%s] The input values didn't produce the expected output", tc.desc))
		}

		assert.Equal(t, tc.res, res, fmt.Sprintf("[%s] The input values didn't produce the expected output", tc.desc))
	}
}

func TestComponentName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc       string
		domainName string
		res        string
		err        error
	}{
		{
			desc:       "one level of subdomains",
			domainName: "foo.bar.com",
			res:        "foo",
		},
		{
			desc:       "two levels of subdomains",
			domainName: "foo.bar.quux.com",
			res:        "foo",
		},
		{
			desc:       "malformed domain",
			domainName: "not a domain name",
			res:        "",
			err:        malformedCloudConfigKeyError,
		},
		{
			desc:       "empty domain",
			domainName: "",
			res:        "",
			err:        malformedCloudConfigKeyError,
		},
	}

	for _, tc := range tests {
		res, err := componentName(tc.domainName)

		if err != nil {
			assert.True(t, IsMalformedCloudConfigKey(err), fmt.Sprintf("[%s] The input values didn't produce the expected output", tc.desc))
		}

		assert.Equal(t, tc.res, res, fmt.Sprintf("[%s] The input values didn't produce the expected output", tc.desc))
	}
}

func Test_VersionBundleVersion(t *testing.T) {
	t.Parallel()
	expectedVersion := "0.1.0"

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			VersionBundle: v1alpha1.AWSConfigSpecVersionBundle{
				Version: "0.1.0",
			},
		},
	}

	if VersionBundleVersion(customObject) != expectedVersion {
		t.Fatalf("Expected version in version bundle to be %s but was %s", expectedVersion, VersionBundleVersion(customObject))
	}
}

func Test_BucketObjectName(t *testing.T) {
	t.Parallel()
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			VersionBundle: v1alpha1.AWSConfigSpecVersionBundle{
				Version: "0.1.0",
			},
		},
	}
	role := "worker"

	e := fmt.Sprintf("version/0.1.0/cloudconfig/%s/worker", CloudConfigVersion)
	a := BucketObjectName(customObject, role)
	if e != a {
		t.Fatalf("expected %s got %s", e, a)
	}
}

func Test_RoleName(t *testing.T) {
	t.Parallel()
	expectedName := "test-cluster-worker-EC2-K8S-Role"
	profileType := "worker"

	cluster := v1alpha1.Cluster{
		ID: "test-cluster",
	}

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: cluster,
		},
	}

	actual := RoleName(customObject, profileType)
	if actual != expectedName {
		t.Fatalf("Expected  name '%s' but was '%s'", expectedName, actual)
	}
}

func Test_PolicyName(t *testing.T) {
	t.Parallel()
	expectedName := "test-cluster-worker-EC2-K8S-Policy"
	profileType := "worker"

	cluster := v1alpha1.Cluster{
		ID: "test-cluster",
	}

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: cluster,
		},
	}

	actual := PolicyName(customObject, profileType)
	if actual != expectedName {
		t.Fatalf("Expected  name '%s' but was '%s'", expectedName, actual)
	}
}

func Test_PeerAccessRoleName(t *testing.T) {
	t.Parallel()
	expectedName := "test-cluster-vpc-peer-access"

	cluster := v1alpha1.Cluster{
		ID: "test-cluster",
	}

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: cluster,
		},
	}

	actual := PeerAccessRoleName(customObject)
	if actual != expectedName {
		t.Fatalf("Expected  name '%s' but was '%s'", expectedName, actual)
	}
}

func Test_MasterCount(t *testing.T) {
	t.Parallel()
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			AWS: v1alpha1.AWSConfigSpecAWS{
				Masters: []v1alpha1.AWSConfigSpecAWSNode{
					{
						InstanceType: "m3.medium",
					},
					{
						InstanceType: "m3.medium",
					},
				},
			},
		},
	}

	expected := 2
	actual := MasterCount(customObject)
	if actual != expected {
		t.Fatalf("Expected master count %d but was %d", expected, actual)
	}
}

func Test_PrivateSubnetCIDR(t *testing.T) {
	t.Parallel()
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			AWS: v1alpha1.AWSConfigSpecAWS{
				VPC: v1alpha1.AWSConfigSpecAWSVPC{
					PrivateSubnetCIDR: "172.31.0.0/16",
				},
			},
		},
	}
	expected := "172.31.0.0/16"
	actual := PrivateSubnetCIDR(customObject)

	if actual != expected {
		t.Fatalf("Expected PrivateSubnetCIDR %s but was %s", expected, actual)
	}
}

func Test_CIDR(t *testing.T) {
	t.Parallel()
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			AWS: v1alpha1.AWSConfigSpecAWS{
				VPC: v1alpha1.AWSConfigSpecAWSVPC{
					CIDR: "172.31.0.0/16",
				},
			},
		},
	}
	expected := "172.31.0.0/16"
	actual := CIDR(customObject)

	if actual != expected {
		t.Fatalf("Expected CIDR %s but was %s", expected, actual)
	}
}

func Test_PeerID(t *testing.T) {
	t.Parallel()
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			AWS: v1alpha1.AWSConfigSpecAWS{
				VPC: v1alpha1.AWSConfigSpecAWSVPC{
					PeerID: "vpc-abcd",
				},
			},
		},
	}
	expected := "vpc-abcd"
	actual := PeerID(customObject)

	if actual != expected {
		t.Fatalf("Expected PeerID %s but was %s", expected, actual)
	}
}

func Test_ImageID(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description     string
		customObject    v1alpha1.AWSConfig
		errorMatcher    func(error) bool
		expectedImageID string
	}{
		{
			description: "basic match",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "eu-central-1",
					},
				},
			},
			errorMatcher:    nil,
			expectedImageID: "ami-0f46c2ed46d8157aa",
		},
		{
			description: "different region",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "eu-west-1",
					},
				},
			},
			errorMatcher:    nil,
			expectedImageID: "ami-0628e483315b5d17e",
		},
		{
			description: "invalid region",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "invalid-1",
					},
				},
			},
			errorMatcher: IsInvalidConfig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			imageID, err := ImageID(tc.customObject)
			if tc.errorMatcher != nil && err == nil {
				t.Error("expected error didn't happen")
			}

			if tc.errorMatcher != nil && !tc.errorMatcher(err) {
				t.Error("expected", true, "got", false)
			}

			if tc.expectedImageID != imageID {
				t.Errorf("unexpected imageID, expecting %q, want %q", tc.expectedImageID, imageID)
			}
		})
	}
}

func Test_TargetLogBucketName(t *testing.T) {
	t.Parallel()
	expectedName := "test-cluster-g8s-access-logs"

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	if TargetLogBucketName(customObject) != expectedName {
		t.Fatalf("Expected target bucket name %s but was %s", expectedName, TargetLogBucketName(customObject))
	}
}

func Test_RegionARN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		region            string
		expectedRegionARN string
		description       string
	}{
		{
			description:       "eu region",
			region:            "eu-central-1",
			expectedRegionARN: "aws",
		},
		{
			description:       "china region",
			region:            "cn-north-1",
			expectedRegionARN: "aws-cn",
		},
		{
			description:       "unknown region",
			region:            "unknown-region-1",
			expectedRegionARN: "aws",
		},
	}

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			AWS: v1alpha1.AWSConfigSpecAWS{
				Region: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			customObject.Spec.AWS.Region = tc.region

			actual := RegionARN(customObject)

			if actual != tc.expectedRegionARN {
				t.Fatalf("Expected region ARN %q but was %q", tc.expectedRegionARN, actual)
			}
		})
	}
}

func Test_MasterRoleARN(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description     string
		customObject    v1alpha1.AWSConfig
		accountID       string
		expectedRoleARN string
	}{
		{
			description: "common partition",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "myclusterid",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "eu-central-1",
					},
				},
			},
			accountID:       "myaccountid",
			expectedRoleARN: "arn:aws:iam::myaccountid:role/myclusterid-master-EC2-K8S-Role",
		},
		{
			description: "china partition",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "myclusterid",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "cn-north-1",
					},
				},
			},
			accountID:       "myaccountid",
			expectedRoleARN: "arn:aws-cn:iam::myaccountid:role/myclusterid-master-EC2-K8S-Role",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			roleARN := MasterRoleARN(tc.customObject, tc.accountID)
			if tc.expectedRoleARN != roleARN {
				t.Errorf("unexpected Master role ARN, expecting %q, want %q", tc.expectedRoleARN, roleARN)
			}
		})
	}
}

func Test_WorkerRoleARN(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description     string
		customObject    v1alpha1.AWSConfig
		accountID       string
		expectedRoleARN string
	}{
		{
			description: "common partition",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "myclusterid",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "eu-central-1",
					},
				},
			},
			accountID:       "myaccountid",
			expectedRoleARN: "arn:aws:iam::myaccountid:role/myclusterid-worker-EC2-K8S-Role",
		},
		{
			description: "china partition",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "myclusterid",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "cn-north-1",
					},
				},
			},
			accountID:       "myaccountid",
			expectedRoleARN: "arn:aws-cn:iam::myaccountid:role/myclusterid-worker-EC2-K8S-Role",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			roleARN := WorkerRoleARN(tc.customObject, tc.accountID)
			if tc.expectedRoleARN != roleARN {
				t.Errorf("unexpected Worker role ARN, expecting %q, want %q", tc.expectedRoleARN, roleARN)
			}
		})
	}
}

func Test_getResourcenameWithTimeHash_Format(t *testing.T) {
	t.Parallel()

	clusterID := "test-cluster"
	prefix := "FooResource"

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: clusterID,
			},
		},
	}

	n := getResourcenameWithTimeHash(prefix, customObject)

	if !strings.HasPrefix(n, prefix) {
		t.Fatalf("expected %s to have prefix %s", n, prefix)
	}
	if strings.Contains(n, "-") {
		t.Fatalf("expected %s to not contain dashes", n)
	}
}
