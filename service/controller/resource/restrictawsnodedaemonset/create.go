package restrictawsnodedaemonset

import (
	"context"
)

const (
	dsNamespace     = "kube-system"
	awsNodeDsName   = "aws-node"
	KubeProxyDsName = "kube-proxy"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	r.logger.Debugf(ctx, "This AWS operator version does not implement this feature.")

	return nil
}
