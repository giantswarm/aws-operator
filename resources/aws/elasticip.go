package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	microerror "github.com/giantswarm/microkit/error"
)

const (
	allocationID = "eipalloc-cc1657a5"
	elasticIP    = "35.158.16.27"
)

type ElasticIP struct {
	InstanceID string
	name       string
	AWSEntity
}

func (e *ElasticIP) CreateIfNotExists() (bool, error) {
	return false, microerror.MaskAny(notImplementedMethodError)
}

func (e *ElasticIP) CreateOrFail() error {
	if _, err := e.Clients.EC2.AssociateAddress(&ec2.AssociateAddressInput{
		AllocationId: aws.String(allocationID),
		InstanceId:   aws.String(e.InstanceID),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	e.name = elasticIP

	return nil
}

func (e *ElasticIP) Delete() error {
	return microerror.MaskAny(notImplementedMethodError)
}

func (e ElasticIP) Name() string {
	return e.name
}
