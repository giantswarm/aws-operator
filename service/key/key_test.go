package key

import (
	"testing"

	"github.com/giantswarm/awstpr"
	awsspec "github.com/giantswarm/awstpr/spec"
	"github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
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
		clusterVersion string
		expectedResult bool
	}{
		{
			clusterVersion: "cloud-formation",
			expectedResult: true,
		},
		{
			clusterVersion: "v_0_1_0",
			expectedResult: false,
		},
		{
			clusterVersion: "",
			expectedResult: false,
		},
	}

	for _, tc := range tests {
		cluster := awstpr.CustomObject{
			Spec: awstpr.Spec{
				Cluster: clustertpr.Spec{
					Version: tc.clusterVersion,
				},
			},
		}

		if UseCloudFormation(cluster) != tc.expectedResult {
			t.Fatalf("Expected use cloud formation to be %t but was %t", tc.expectedResult, UseCloudFormation(cluster))
		}
	}
}
