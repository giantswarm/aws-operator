package create

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/awstpr"
	microerror "github.com/giantswarm/microkit/error"
)

type recordSetInput struct {
	Cluster      awstpr.CustomObject
	Client       *route53.Route53
	Resource     resources.DNSNamedResource
	Domain       string
	HostedZoneID string
}

func (s *Service) deleteRecordSet(input recordSetInput) error {
	rs := &awsresources.RecordSet{
		Client:       input.Client,
		Resource:     input.Resource,
		Domain:       input.Domain,
		HostedZoneID: input.HostedZoneID,
	}

	if err := rs.Delete(); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (s *Service) createRecordSet(input recordSetInput) error {
	// Create DNS records for LB.
	apiRecordSet := &awsresources.RecordSet{
		Client:       input.Client,
		Resource:     input.Resource,
		Domain:       input.Domain,
		HostedZoneID: input.HostedZoneID,
	}

	if err := apiRecordSet.CreateOrFail(); err != nil {
		return microerror.MaskAnyf(err, "error registering DNS record '%s'", apiRecordSet.Domain)
	}

	s.logger.Log("debug", fmt.Sprintf("created or reused DNS record '%s'", apiRecordSet.Domain))

	return nil
}
