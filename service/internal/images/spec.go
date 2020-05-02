package images

import (
	"context"

	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/v_6_0_0"
)

type Interface interface {
	// ForRelease looks up necessary information to compute the relevant Cloud
	// Config images for the given object's release version. Paramter obj must be
	// a metav1.Object and contain the Giant Swarm specific release version label.
	ForRelease(ctx context.Context, obj interface{}) (k8scloudconfig.Images, error)
}
