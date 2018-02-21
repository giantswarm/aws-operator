package ebsvolume

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/awsconfig/v5/key"
)

const (
	cloudProviderClusterTagValue        = "owned"
	cloudProviderPersistentVolumeTagKey = "kubernetes.io/created-for/pv/name"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	currentState, err := r.persistentVolumes(customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return currentState, nil
}

func (r *Resource) persistentVolumes(customObject v1alpha1.AWSConfig) (*EBSVolumeState, error) {
	volumeState := &EBSVolumeState{}
	volumeIDs := []string{}

	clusterTag := key.ClusterCloudProviderTag(customObject)
	filters := []*ec2.Filter{
		{
			Name: aws.String(fmt.Sprintf("tag:%s", clusterTag)),
			Values: []*string{
				aws.String(cloudProviderClusterTagValue),
			},
		},
	}
	output, err := r.clients.EC2.DescribeVolumes(&ec2.DescribeVolumesInput{
		Filters: filters,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, vol := range output.Volumes {
		if containsPersistentVolumeTag(vol.Tags) {
			volumeIDs = append(volumeIDs, *vol.VolumeId)
		}
	}

	volumeState.VolumeIDs = volumeIDs

	return volumeState, nil
}

func containsPersistentVolumeTag(tags []*ec2.Tag) bool {
	for _, tag := range tags {
		if *tag.Key == cloudProviderPersistentVolumeTagKey {
			return true
		}
	}

	return false
}
