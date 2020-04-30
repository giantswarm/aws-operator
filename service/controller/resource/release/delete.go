package release

import (
	"context"
)

func (r *Resource) EnsureDeleted(_ context.Context, _ interface{}) error {
	return nil
}
