package lifecycle

import (
	"context"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
