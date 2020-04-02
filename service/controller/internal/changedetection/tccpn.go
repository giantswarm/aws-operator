package changedetection

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type TCCPNConfig struct {
	Logger micrologger.Logger
}

// TCCPN is a detection service implementation deciding if a control plane should be
// updated.
type TCCPN struct {
	logger micrologger.Logger
}

func NewTCCPN(config TCCPNConfig) (*TCCPN, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	t := &TCCPN{
		logger: config.Logger,
	}

	return t, nil
}

// ShouldUpdate determines whether the reconciled tenant cluster control plane
// should be updated. A tenant cluster control plane is only allowed to update in
// the following cases.
//
// TODO
//
func (t *TCCPN) ShouldUpdate(ctx context.Context, md infrastructurev1alpha2.AWSControlPlane) (bool, error) {

	return false, nil
}
