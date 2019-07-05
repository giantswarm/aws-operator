package ebs

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/micrologger/microloggertest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/label"
)

func Test_ListVolumes(t *testing.T) {
	t.Parallel()

	customObject := v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				label.Cluster: "test-cluster",
			},
		},
	}

	testCases := []struct {
		description     string
		obj             v1alpha1.Cluster
		filterFuncs     []func(t *ec2.Tag) bool
		expectedVolumes []Volume
		ebsVolumes      []ebsVolumeMock
	}{
		{
			description: "case 0: basic match with no volumes",
			obj:         customObject,
			filterFuncs: []func(t *ec2.Tag) bool{
				NewEtcdVolumeFilter(customObject),
				NewPersistentVolumeFilter(customObject),
			},
			expectedVolumes: nil,
		},
		{
			description: "case 1: basic match with pv volume",
			obj:         customObject,
			filterFuncs: []func(t *ec2.Tag) bool{
				NewEtcdVolumeFilter(customObject),
				NewPersistentVolumeFilter(customObject),
			},
			expectedVolumes: []Volume{
				{
					Attachments: []VolumeAttachment{},
					VolumeID:    "vol-1234",
				},
			},
			ebsVolumes: []ebsVolumeMock{
				{
					volumeID: "vol-1234",
					tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-1234"),
						},
					},
				},
			},
		},
		{
			description: "case 2: basic match with etcd and multiple pv volumes",
			obj:         customObject,
			filterFuncs: []func(t *ec2.Tag) bool{
				NewEtcdVolumeFilter(customObject),
				NewPersistentVolumeFilter(customObject),
			},
			expectedVolumes: []Volume{
				{
					Attachments: []VolumeAttachment{},
					VolumeID:    "vol-1234",
				},
				{
					Attachments: []VolumeAttachment{},
					VolumeID:    "vol-5678",
				},
				{
					Attachments: []VolumeAttachment{},
					VolumeID:    "vol-6789",
				},
			},
			ebsVolumes: []ebsVolumeMock{
				{
					volumeID: "vol-1234",
					tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-1234"),
						},
					},
				},
				{
					volumeID: "vol-5678",
					tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-5678"),
						},
					},
				},
				{
					volumeID: "vol-6789",
					tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("Name"),
							Value: aws.String("test-cluster-etcd"),
						},
					},
				},
			},
		},
		{
			description: "case 3: no match due to cluster tag",
			obj:         customObject,
			filterFuncs: []func(t *ec2.Tag) bool{
				NewEtcdVolumeFilter(customObject),
				NewPersistentVolumeFilter(customObject),
			},
			expectedVolumes: nil,
			ebsVolumes: []ebsVolumeMock{
				{
					volumeID: "vol-1234",
					tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/other-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-1234"),
						},
					},
				},
			},
		},
		{
			description: "case 4: no match due to missing pv tag",
			obj:         customObject,
			filterFuncs: []func(t *ec2.Tag) bool{
				NewEtcdVolumeFilter(customObject),
				NewPersistentVolumeFilter(customObject),
			},
			expectedVolumes: nil,
			ebsVolumes: []ebsVolumeMock{
				{
					volumeID: "vol-1234",
					tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
					},
				},
			},
		},
		{
			description: "case 5: multiple ebs volumes with attachments",
			obj:         customObject,
			filterFuncs: []func(t *ec2.Tag) bool{
				NewEtcdVolumeFilter(customObject),
				NewPersistentVolumeFilter(customObject),
			},
			expectedVolumes: []Volume{
				{
					Attachments: []VolumeAttachment{
						{
							InstanceID: "i-12345",
							Device:     "/dev/sdh",
						},
					},
					VolumeID: "vol-1234",
				},
				{
					Attachments: []VolumeAttachment{
						{
							InstanceID: "i-56789",
							Device:     "/dev/sdh",
						},
					},
					VolumeID: "vol-5678",
				},
			},
			ebsVolumes: []ebsVolumeMock{
				{
					volumeID: "vol-1234",
					attachments: []*ec2.VolumeAttachment{
						{
							Device:     aws.String("/dev/sdh"),
							InstanceId: aws.String("i-12345"),
						},
					},
					tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-1234"),
						},
					},
				},
				{
					volumeID: "vol-5678",
					attachments: []*ec2.VolumeAttachment{
						{
							Device:     aws.String("/dev/sdh"),
							InstanceId: aws.String("i-56789"),
						},
					},
					tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-5678"),
						},
					},
				},
			},
		},
		{
			description: "case 6: only etcd volume",
			obj:         customObject,
			filterFuncs: []func(t *ec2.Tag) bool{
				NewEtcdVolumeFilter(customObject),
			},
			expectedVolumes: []Volume{
				{
					Attachments: []VolumeAttachment{},
					VolumeID:    "vol-6789",
				},
			},
			ebsVolumes: []ebsVolumeMock{
				{
					volumeID: "vol-1234",
					tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-1234"),
						},
					},
				},
				{
					volumeID: "vol-6789",
					tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("Name"),
							Value: aws.String("test-cluster-etcd"),
						},
					},
				},
			},
		},
		{
			description: "case 7: only pv volume",
			obj:         customObject,
			filterFuncs: []func(t *ec2.Tag) bool{
				NewPersistentVolumeFilter(customObject),
			},
			expectedVolumes: []Volume{
				{
					Attachments: []VolumeAttachment{},
					VolumeID:    "vol-1234",
				},
			},
			ebsVolumes: []ebsVolumeMock{
				{
					volumeID: "vol-1234",
					tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-1234"),
						},
					},
				},
				{
					volumeID: "vol-6789",
					tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("Name"),
							Value: aws.String("test-cluster-etcd"),
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			c := Config{
				Client: &EC2ClientMock{
					customObject: tc.obj,
					ebsVolumes:   tc.ebsVolumes,
				},
				Logger: microloggertest.New(),
			}
			e, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			result, err := e.ListVolumes(tc.obj, tc.filterFuncs...)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			if !reflect.DeepEqual(result, tc.expectedVolumes) {
				t.Fatalf("expected volumes '%#v', got '%#v'", tc.expectedVolumes, result)
			}
		})
	}
}
