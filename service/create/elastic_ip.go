package create

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	microerror "github.com/giantswarm/microkit/error"
)

const (
	allocationID = "eipalloc-cc1657a5"
	elasticIP    = "35.158.16.27"
)

func (s *Service) associateElasticIP(svc *ec2.EC2, instanceID string) error {
	params := &ec2.AssociateAddressInput{
		AllocationId: aws.String(allocationID),
		InstanceId:   aws.String(instanceID),
	}

	if _, err := svc.AssociateAddress(params); err != nil {
		return microerror.MaskAny(err)
	}

	s.logger.Log("info", fmt.Sprintf("attached ip %v to instance %v", elasticIP, instanceID))

	return nil
}
