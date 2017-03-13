package create

import (
	"context"

	"github.com/ericchiang/k8s"
	"github.com/ericchiang/k8s/api/v1"
	"github.com/ericchiang/k8s/apis/extensions/v1beta1"
	"github.com/giantswarm/awstpr"
	microerror "github.com/giantswarm/microkit/error"
)

func (s *Service) createTPR() error {
	tpr := &v1beta1.ThirdPartyResource{
		Metadata: &v1.ObjectMeta{
			Name: k8s.String(awstpr.Name),
		},
		Versions: []*v1beta1.APIVersion{
			{Name: k8s.String("v1")},
		},
		Description: k8s.String("Managed Kubernetes on AWS clusters"),
	}
	if _, err := s.k8sClient.ExtensionsV1Beta1().CreateThirdPartyResource(context.Background(), tpr); err != nil {
		return microerror.MaskAny(err)
	}
	return nil
}
