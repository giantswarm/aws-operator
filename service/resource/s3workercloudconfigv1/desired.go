package s3workercloudconfigv1

import "context"

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	return BucketObject{}, nil
}
