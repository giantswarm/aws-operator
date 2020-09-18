package images

import (
	"context"

	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v7/pkg/template"
)

type Interface interface {
	// AMI looks up necessary information to compute the relevant EC2 AMI for the
	// given object's region and release version. Paramter obj must be a
	// metav1.Object and contain the Giant Swarm specific cluster ID label and
	// release version label.
	AMI(ctx context.Context, obj interface{}) (string, error)
	// AWSCNI looks up aws-cni version to compute the relevant Cloud Config
	// images for the given object's release version. Paramter obj must be a
	// metav1.Object and contain the Giant Swarm specific release version label.
	AWSCNI(ctx context.Context, obj interface{}) (string, error)
	// CC looks up necessary information to compute the relevant Cloud Config
	// images for the given object's release version. Paramter obj must be a
	// metav1.Object and contain the Giant Swarm specific release version label.
	CC(ctx context.Context, obj interface{}) (k8scloudconfig.Images, error)
	// Versions looks up necessary information to compute the relevant Cloud Config
	// images versions for the given object's release version. Paramter obj must be a
	// metav1.Object and contain the Giant Swarm specific release version label.
	Versions(ctx context.Context, obj interface{}) (k8scloudconfig.Versions, error)
}
