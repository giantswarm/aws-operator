package keyv2

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/stretchr/testify/assert"
)

func Test_AutoScalingGroupName(t *testing.T) {
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

func Test_BucketName(t *testing.T) {
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

func Test_ClusterNamespace(t *testing.T) {
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

func Test_ClusterTags(t *testing.T) {
	expectedID := "test-cluster"
	expectedTags := map[string]string{
		"kubernetes.io/cluster/test-cluster": "owned",
	}

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: expectedID,
			},
		},
	}

	if !reflect.DeepEqual(expectedTags, ClusterTags(customObject)) {
		t.Fatalf("Expected cluster tags %v but was %v", expectedTags, ClusterTags(customObject))
	}
}

func Test_ClusterVersion(t *testing.T) {
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

func Test_HasClusterVersion(t *testing.T) {
	tests := []struct {
		customObject   v1alpha1.AWSConfig
		expectedResult bool
	}{
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{},
			},
			expectedResult: false,
		},
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						Version: "",
					},
				},
			},
			expectedResult: false,
		},
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						Version: "v_0_1_0",
					},
				},
			},
			expectedResult: true,
		},
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						Version: "v_0_2_0",
					},
				},
			},
			expectedResult: false,
		},
	}

	for _, tc := range tests {
		if HasClusterVersion(tc.customObject) != tc.expectedResult {
			t.Fatalf("Expected has cluster version to be %t but was %t", tc.expectedResult, HasClusterVersion(tc.customObject))
		}
	}
}

func Test_IngressControllerInsecurePort(t *testing.T) {
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

func Test_KubernetesAPISecurePort(t *testing.T) {
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
	tests := []struct {
		customObject    v1alpha1.AWSConfig
		expectedImageID string
	}{
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							v1alpha1.AWSConfigSpecAWSNode{
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

func Test_MasterInstanceType(t *testing.T) {
	tests := []struct {
		customObject         v1alpha1.AWSConfig
		expectedInstanceType string
	}{
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Masters: []v1alpha1.AWSConfigSpecAWSNode{
							v1alpha1.AWSConfigSpecAWSNode{
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
							v1alpha1.AWSConfigSpecAWSNode{},
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

func Test_RouteTableName(t *testing.T) {
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
func Test_SecurityGroupName(t *testing.T) {
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
	expectedCount := 2

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			AWS: v1alpha1.AWSConfigSpecAWS{
				Workers: []v1alpha1.AWSConfigSpecAWSNode{
					v1alpha1.AWSConfigSpecAWSNode{
						InstanceType: "m3.medium",
					},
					v1alpha1.AWSConfigSpecAWSNode{
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

func Test_WorkerImageID(t *testing.T) {
	tests := []struct {
		customObject    v1alpha1.AWSConfig
		expectedImageID string
	}{
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							v1alpha1.AWSConfigSpecAWSNode{
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
	tests := []struct {
		customObject         v1alpha1.AWSConfig
		expectedInstanceType string
	}{
		{
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							v1alpha1.AWSConfigSpecAWSNode{
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

func Test_UseCloudFormation(t *testing.T) {
	tests := []struct {
		versionBundleVersion string
		expectedResult       bool
	}{
		{
			versionBundleVersion: "0.1.0",
			expectedResult:       false,
		},
		{
			versionBundleVersion: "0.2.0",
			expectedResult:       true,
		},
		{
			versionBundleVersion: "",
			expectedResult:       false,
		},
	}

	for _, tc := range tests {
		cluster := v1alpha1.AWSConfig{
			Spec: v1alpha1.AWSConfigSpec{
				VersionBundle: v1alpha1.AWSConfigSpecVersionBundle{
					Version: tc.versionBundleVersion,
				},
			},
		}

		if UseCloudFormation(cluster) != tc.expectedResult {
			t.Fatalf("Expected use cloud formation to be %t but was %t", tc.expectedResult, UseCloudFormation(cluster))
		}
	}
}

func Test_InstanceProfileName(t *testing.T) {
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

func TestRootDir(t *testing.T) {
	testCases := []struct {
		desc          string
		rootElement   string
		baseDir       string
		expectedDir   string
		expectedError bool
	}{
		{
			desc:          "basic case, one level",
			rootElement:   "aws-operator",
			baseDir:       "/home/user/aws-operator/dir",
			expectedDir:   "/home/user/aws-operator",
			expectedError: false,
		},
		{
			desc:          "basic case, two levels",
			rootElement:   "aws-operator",
			baseDir:       "/home/user/aws-operator/dir/subdir",
			expectedDir:   "/home/user/aws-operator",
			expectedError: false,
		},
		{
			desc:          "aws-operator as first dir",
			rootElement:   "aws-operator",
			baseDir:       "/aws-operator/dir/subdir",
			expectedDir:   "/aws-operator",
			expectedError: false,
		},
		{
			desc:          "aws-operator not present",
			rootElement:   "aws-operator",
			baseDir:       "/home/user/dir/subdir",
			expectedDir:   "",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			actual, err := RootDir(tc.baseDir, tc.rootElement)

			if err != nil && !tc.expectedError {
				t.Errorf("unexpected error: %v", err)
			}

			if actual != tc.expectedDir {
				t.Errorf("unexpected result, want %q, got %q", tc.expectedDir, actual)
			}
		})
	}
}

func Test_VersionBundleVersion(t *testing.T) {
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
	version := "v_0_1_0"
	suffix := "mysuffix"

	expectedBucketObjectName := "cloudconfig/v_0_1_0/mysuffix"
	actualBucketObjectName := BucketObjectName(version, suffix)
	if expectedBucketObjectName != actualBucketObjectName {
		t.Fatalf("Expected bucket object name %q but was %q", expectedBucketObjectName, actualBucketObjectName)
	}
}

func Test_RoleName(t *testing.T) {
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
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			AWS: v1alpha1.AWSConfigSpecAWS{
				Masters: []v1alpha1.AWSConfigSpecAWSNode{
					v1alpha1.AWSConfigSpecAWSNode{
						InstanceType: "m3.medium",
					},
					v1alpha1.AWSConfigSpecAWSNode{
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
