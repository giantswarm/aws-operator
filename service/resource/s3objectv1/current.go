package s3objectv1

import (
	"context"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	return BucketObject{}, nil
}
