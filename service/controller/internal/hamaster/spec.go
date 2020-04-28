package hamaster

import "context"

type Interface interface {
	Enabled(ctx context.Context, cluster string) (bool, error)
}
