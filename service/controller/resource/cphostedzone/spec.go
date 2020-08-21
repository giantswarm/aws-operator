package cphostedzone

import (
	"github.com/aws/aws-sdk-go/service/route53"
)

type Route53 interface {
	ListHostedZonesByName(input *route53.ListHostedZonesByNameInput) (*route53.ListHostedZonesByNameOutput, error)
}
