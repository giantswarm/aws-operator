package prepareawscniformigration

import (
	"context"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	r.logger.Debugf(ctx, "This AWS operator version does not implement this feature.")

	return nil
}
