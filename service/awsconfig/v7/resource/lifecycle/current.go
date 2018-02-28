package lifecycle

import (
	"context"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	r.logger.LogCtx(ctx, "level", "debug", "message", "lifecycle resource got executed")
	return nil, nil
}
