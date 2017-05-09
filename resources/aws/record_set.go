package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/giantswarm/aws-operator/resources"
	microerror "github.com/giantswarm/microkit/error"
)

type RecordSet struct {
	// Domain is the domain name for the record.
	Domain string
	// HostedZoneID is the ID of the Hosted Zone the record should be created in.
	HostedZoneID string
	// Client is the AWS client.
	Client *route53.Route53
	// Resource is the AWS resource the record should be created for.
	Resource resources.DNSNamedResource
}

// CreateIfNotExists is not implemented because AWS provides UPSERT functionality for DNS records
func (record RecordSet) CreateIfNotExists() (bool, error) {
	return false, microerror.MaskAny(notImplementedMethodError)
}

func (record RecordSet) CreateOrFail() error {
	if err := record.perform(route53.ChangeActionUpsert); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (record RecordSet) Delete() error {
	if err := record.perform(route53.ChangeActionDelete); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (record RecordSet) perform(action string) error {
	if record.Client == nil {
		return clientNotInitializedError
	}

	params := record.buildParams(action)

	if _, err := record.Client.ChangeResourceRecordSets(params); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (record RecordSet) buildParams(action string) *route53.ChangeResourceRecordSetsInput {
	return &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(action),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(record.Domain),
						Type: aws.String(route53.RRTypeA),
						AliasTarget: &route53.AliasTarget{
							HostedZoneId:         aws.String(record.Resource.HostedZoneID()),
							DNSName:              aws.String(record.Resource.DNSName()),
							EvaluateTargetHealth: aws.Bool(false),
						},
					},
				},
			},
		},
		HostedZoneId: aws.String(record.HostedZoneID),
	}
}
