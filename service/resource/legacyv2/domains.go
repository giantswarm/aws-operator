package legacyv1

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"
)

type recordSetInput struct {
	Cluster      awstpr.CustomObject
	Client       *route53.Route53
	Resource     resources.DNSNamedResource
	Value        string
	Domain       string
	HostedZoneID string
	Type         string
}

func (s *Resource) deleteRecordSet(input recordSetInput) error {
	rs := &awsresources.RecordSet{
		Client:       input.Client,
		Resource:     input.Resource,
		Value:        input.Value,
		Domain:       input.Domain,
		HostedZoneID: input.HostedZoneID,
		Type:         input.Type,
	}

	if err := rs.Delete(); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (s *Resource) createRecordSet(input recordSetInput) error {
	// Create DNS records for LB.
	apiRecordSet := &awsresources.RecordSet{
		Client:       input.Client,
		Resource:     input.Resource,
		Value:        input.Value,
		Domain:       input.Domain,
		HostedZoneID: input.HostedZoneID,
		Type:         input.Type,
	}

	if err := apiRecordSet.CreateOrFail(); err != nil {
		return microerror.Maskf(err, "error registering DNS record '%s'", apiRecordSet.Domain)
	}

	s.logger.Log("debug", fmt.Sprintf("created or reused DNS record '%s'", apiRecordSet.Domain))

	return nil
}
