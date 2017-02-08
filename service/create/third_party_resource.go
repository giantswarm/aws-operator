package create

import (
	"github.com/giantswarm/awstpr"
	microerror "github.com/giantswarm/microkit/error"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func (s *Service) createTPR() error {
	tpr := &v1beta1.ThirdPartyResource{
		ObjectMeta: v1.ObjectMeta{
			Name: awstpr.Name,
		},
		Versions: []v1beta1.APIVersion{
			{Name: "v1"},
		},
		Description: "Managed Kubernetes on AWS clusters",
	}
	_, err := s.k8sClient.Extensions().ThirdPartyResources().Create(tpr)
	if err != nil && !errors.IsAlreadyExists(err) {
		return microerror.MaskAny(err)
	}
	return nil
}
