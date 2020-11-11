package releases

import (
	"context"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
)

type Interface interface {
	// Release returns the release object from a certain version
	Release(ctx context.Context, version string) (releasev1alpha1.Release, error)
}
