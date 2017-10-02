package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/giantswarm/aws-operator/resources"
	"github.com/giantswarm/microerror"
)

const (
	// Default TTL for CNAME domains.
	defaultTTL int64 = 900
)

type RecordSet struct {
	// Domain is the domain name for the record.
	Domain string
	// HostedZoneID is the ID of the Hosted Zone the record should be created in.
	HostedZoneID string
	Type         string
	// Client is the AWS client.
	Client *route53.Route53
	// Resource is the AWS resource the record should be created for.
	Resource resources.DNSNamedResource
	Value    string
}

// CreateIfNotExists is not implemented because AWS provides UPSERT functionality for DNS records
func (record RecordSet) CreateIfNotExists() (bool, error) {
	return false, microerror.Mask(notImplementedMethodError)
}

func (record RecordSet) CreateOrFail() error {
	if err := record.perform(route53.ChangeActionUpsert); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (record RecordSet) Delete() error {
	if err := record.perform(route53.ChangeActionDelete); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (record RecordSet) perform(action string) error {
	if record.Client == nil {
		return clientNotInitializedError
	}

	params := record.buildParams(action)

	if _, err := record.Client.ChangeResourceRecordSets(params); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (record RecordSet) buildParams(action string) *route53.ChangeResourceRecordSetsInput {
	var aliasTarget *route53.AliasTarget
	var resourceRecords []*route53.ResourceRecord
	var ttl *int64

	fmt.Printf("record.Resource.HostedZoneID(): %#v\n", record.Resource.HostedZoneID())
	fmt.Printf("record.Resource.DNSName(): %#v\n", record.Resource.DNSName())
	fmt.Printf("aws.String(record.Value): %#v\n", aws.String(record.Value))

	switch record.Type {
	case route53.RRTypeA:
		fmt.Printf("1\n")
		aliasTarget = &route53.AliasTarget{
			HostedZoneId:         aws.String(record.Resource.HostedZoneID()),
			DNSName:              aws.String(record.Resource.DNSName()),
			EvaluateTargetHealth: aws.Bool(false),
		}
	case route53.RRTypeCname:
		fmt.Printf("2\n")
		resourceRecords = append(resourceRecords, &route53.ResourceRecord{
			Value: aws.String(record.Value),
		})
		ttl = aws.Int64(defaultTTL)
	}

	fmt.Printf("aliasTarget: %#v\n", aliasTarget)
	fmt.Printf("resourceRecords: %#v\n", resourceRecords)
	fmt.Printf("record.HostedZoneID: %#v\n", record.HostedZoneID)

	return &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(action),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            aws.String(record.Domain),
						Type:            aws.String(record.Type),
						AliasTarget:     aliasTarget,
						ResourceRecords: resourceRecords,
						TTL:             ttl,
					},
				},
			},
		},
		HostedZoneId: aws.String(record.HostedZoneID),
	}
}
