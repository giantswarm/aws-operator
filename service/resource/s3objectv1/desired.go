package s3objectv1

import "context"

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	return BucketObject{}, nil
}
