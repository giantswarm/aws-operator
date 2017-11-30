package key

import (
	"fmt"
	"testing"

	"github.com/giantswarm/awstpr"
	awsspec "github.com/giantswarm/awstpr/spec"
	"github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/microerror"
	"github.com/stretchr/testify/assert"
)

func Test_AutoScalingGroupName(t *testing.T) {
	expectedName := "test-cluster-worker"
	groupName := "worker"

	cluster := clustertpr.Spec{
		Cluster: spec.Cluster{
			ID: "test-cluster",
		},
		Customer: spec.Customer{
			ID: "test-customer",
		},
	}

	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: cluster,
		},
	}

	if AutoScalingGroupName(customObject, groupName) != expectedName {
		t.Fatalf("Expected auto scaling group name %s but was %s", expectedName, AutoScalingGroupName(customObject, groupName))
	}
}

func Test_AvailabilityZone(t *testing.T) {
	expectedAZ := "eu-central-1a"

	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			AWS: awsspec.AWS{
				AZ: "eu-central-1a",
			},
		},
	}

	if AvailabilityZone(customObject) != expectedAZ {
		t.Fatalf("Expected availability zone %s but was %s", expectedAZ, AvailabilityZone(customObject))
	}
}

func Test_ClusterID(t *testing.T) {
	expectedID := "test-cluster"

	cluster := clustertpr.Spec{
		Cluster: spec.Cluster{
			ID: expectedID,
		},
		Customer: spec.Customer{
			ID: "test-customer",
		},
	}

	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: cluster,
		},
	}

	if ClusterID(customObject) != expectedID {
		t.Fatalf("Expected cluster ID %s but was %s", expectedID, ClusterID(customObject))
	}
}

func Test_ClusterCustomer(t *testing.T) {
	expectedCustomer := "test-customer"

	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: clustertpr.Spec{
				Cluster: spec.Cluster{
					ID: "test-cluster",
				},
				Customer: spec.Customer{
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

	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: clustertpr.Spec{
				Cluster: spec.Cluster{
					ID: expectedID,
				},
			},
		},
	}

	if ClusterNamespace(customObject) != expectedID {
		t.Fatalf("Expected cluster ID %s but was %s", expectedID, ClusterNamespace(customObject))
	}
}

func Test_ClusterVersion(t *testing.T) {
	expectedVersion := "v_0_1_0"

	cluster := clustertpr.Spec{
		Version: expectedVersion,
	}

	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: cluster,
		},
	}

	if ClusterVersion(customObject) != expectedVersion {
		t.Fatalf("Expected cluster version %s but was %s", expectedVersion, ClusterVersion(customObject))
	}
}

func Test_HasClusterVersion(t *testing.T) {
	tests := []struct {
		customObject   awstpr.CustomObject
		expectedResult bool
	}{
		{
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{},
			},
			expectedResult: false,
		},
		{
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Version: "",
					},
				},
			},
			expectedResult: false,
		},
		{
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Version: "v_0_1_0",
					},
				},
			},
			expectedResult: true,
		},
		{
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
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

func Test_MasterImageID(t *testing.T) {
	tests := []struct {
		customObject    awstpr.CustomObject
		expectedImageID string
	}{
		{
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					AWS: awsspec.AWS{
						Masters: []aws.Node{
							aws.Node{
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
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{},
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

func Test_MasterInstanceType(t *testing.T) {
	tests := []struct {
		customObject         awstpr.CustomObject
		expectedInstanceType string
	}{
		{
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					AWS: awsspec.AWS{
						Masters: []aws.Node{
							aws.Node{
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
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					AWS: awsspec.AWS{
						Masters: []aws.Node{
							aws.Node{},
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

	cluster := clustertpr.Spec{
		Cluster: spec.Cluster{
			ID: "test-cluster",
		},
	}

	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
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

	cluster := clustertpr.Spec{
		Cluster: spec.Cluster{
			ID: "test-cluster",
		},
		Customer: spec.Customer{
			ID: "test-customer",
		},
	}

	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
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

	cluster := clustertpr.Spec{
		Cluster: spec.Cluster{
			ID: "test-cluster",
		},
	}

	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: cluster,
		},
	}

	if SubnetName(customObject, suffix) != expectedName {
		t.Fatalf("Expected subnet name %s but was %s", expectedName, SubnetName(customObject, suffix))
	}

}

func Test_WorkerCount(t *testing.T) {
	expectedCount := 2

	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			AWS: awsspec.AWS{
				Workers: []aws.Node{
					aws.Node{
						InstanceType: "m3.medium",
					},
					aws.Node{
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
		customObject    awstpr.CustomObject
		expectedImageID string
	}{
		{
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					AWS: awsspec.AWS{
						Workers: []aws.Node{
							aws.Node{
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
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					AWS: awsspec.AWS{
						Workers: []aws.Node{},
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
		customObject         awstpr.CustomObject
		expectedInstanceType string
	}{
		{
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					AWS: awsspec.AWS{
						Workers: []aws.Node{
							aws.Node{
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
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					AWS: awsspec.AWS{
						Workers: []aws.Node{},
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

func Test_MainStackName(t *testing.T) {
	expected := "xyz-main"

	cluster := awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: clustertpr.Spec{
				Cluster: spec.Cluster{
					ID: "xyz",
				},
			},
		},
	}

	actual := MainStackName(cluster)
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
		cluster := awstpr.CustomObject{
			Spec: awstpr.Spec{
				VersionBundle: awsspec.VersionBundle{
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

	cluster := clustertpr.Spec{
		Cluster: spec.Cluster{
			ID: "test-cluster",
		},
	}

	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
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
		tpo        awstpr.CustomObject
		res        string
		err        error
	}{
		{
			desc:       "works",
			domainName: "component.foo.bar.example.com",
			tpo: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "foo-customer",
						},
					},
				},
			},
			res: "foo-customer-component",
		},
		{
			desc:       "also works",
			domainName: "component.of.a.well.formed.domain",
			tpo: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "quux-the-customer",
						},
					},
				},
			},
			res: "quux-the-customer-component",
		},
		{
			desc:       "missing ID key in cloudconfig",
			domainName: "component.foo.bar.example.com",
			tpo: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "",
						},
					},
				},
			},
			res: "",
			err: missingCloudConfigKeyError,
		},
		{
			desc:       "malformed domain name",
			domainName: "not a domain name",
			tpo: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "foo-customer",
						},
					},
				},
			},
			res: "",
			err: malformedCloudConfigKeyError,
		},
		{
			desc:       "missing domain name",
			domainName: "",
			tpo: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "foo-customer",
						},
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

	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			VersionBundle: awsspec.VersionBundle{
				Version: "0.1.0",
			},
		},
	}

	if VersionBundleVersion(customObject) != expectedVersion {
		t.Fatalf("Expected version in version bundle to be %s but was %s", expectedVersion, VersionBundleVersion(customObject))
	}
}
