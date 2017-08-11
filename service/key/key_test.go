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

func Test_MasterImageID(t *testing.T) {
	expectedImageID := "ami-d60ad6b9"

	customObject := awstpr.CustomObject{
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
	}

	if MasterImageID(customObject) != expectedImageID {
		t.Fatalf("Expected master image ID %s but was %s", expectedImageID, MasterImageID(customObject))
	}
}

func Test_MasterInstanceType(t *testing.T) {
	expectedInstanceType := "m3.medium"

	customObject := awstpr.CustomObject{
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
	}

	if MasterInstanceType(customObject) != expectedInstanceType {
		t.Fatalf("Expected master instance type %s but was %s", expectedInstanceType, MasterInstanceType(customObject))
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
	expectedImageID := "ami-d60ad6b9"

	customObject := awstpr.CustomObject{
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
	}

	if WorkerImageID(customObject) != expectedImageID {
		t.Fatalf("Expected worker image ID %s but was %s", expectedImageID, WorkerImageID(customObject))
	}
}

func Test_WorkerInstanceType(t *testing.T) {
	expectedInstanceType := "m3.medium"

	customObject := awstpr.CustomObject{
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
	}

	if WorkerInstanceType(customObject) != expectedInstanceType {
		t.Fatalf("Expected worker instance type %s but was %s", expectedInstanceType, WorkerInstanceType(customObject))
	}
}
