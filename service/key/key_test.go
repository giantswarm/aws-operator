package key

import (
	"testing"

	"github.com/giantswarm/awstpr"
	awsspec "github.com/giantswarm/awstpr/spec"
	"github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
)

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
