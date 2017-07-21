package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	microerror "github.com/giantswarm/microkit/error"
)

// SecurityGroupRules allows AWS security group rules to be deleted. Any rules
// referencing other security groups must be deleted before the group can be
// deleted.
type SecurityGroupRules struct {
	Description string
	GroupName   string
	AWSEntity
}

func (s SecurityGroupRules) findExisting() (*ec2.SecurityGroup, error) {
	securityGroups, err := s.Clients.EC2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String(subnetDescription),
				Values: []*string{
					aws.String(s.Description),
				},
			},
			&ec2.Filter{
				Name: aws.String(subnetGroupName),
				Values: []*string{
					aws.String(s.GroupName),
				},
			},
		},
	})
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	if len(securityGroups.SecurityGroups) < 1 {
		return nil, microerror.MaskAnyf(notFoundError, notFoundErrorFormat, SecurityGroupType, s.GroupName)
	} else if len(securityGroups.SecurityGroups) > 1 {
		return nil, microerror.MaskAny(tooManyResultsError)
	}

	return securityGroups.SecurityGroups[0], nil
}

// Delete deletes any security group rules that reference other groups.
// This must happen before the security group can be deleted. Rules using a
// CIDR do not need to be deleted.
func (s SecurityGroupRules) Delete() error {
	securityGroup, err := s.findExisting()
	if err != nil {
		return microerror.MaskAny(err)
	}

	var params *ec2.RevokeSecurityGroupIngressInput
	params = &ec2.RevokeSecurityGroupIngressInput{
		GroupId:       securityGroup.GroupId,
		IpPermissions: securityGroup.IpPermissions,
	}
	if _, err := s.Clients.EC2.RevokeSecurityGroupIngress(params); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}
