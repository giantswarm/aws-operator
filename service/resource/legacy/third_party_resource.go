package legacy

import (
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func (s *Service) createTPR() error {
	tpr := &v1beta1.ThirdPartyResource{
		ObjectMeta: apismetav1.ObjectMeta{
			Name: awstpr.Name,
		},
		Versions: []v1beta1.APIVersion{
			{Name: "v1"},
		},
		Description: "Managed Kubernetes on AWS clusters",
	}
	_, err := s.k8sClient.Extensions().ThirdPartyResources().Create(tpr)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return microerror.Mask(err)
	}
	return nil
}
