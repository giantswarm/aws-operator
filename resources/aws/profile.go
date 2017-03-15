package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	microerror "github.com/giantswarm/microkit/error"
)

type InstanceProfile struct {
	ClusterID string
	name      string
	AWSEntity
}

func (ip *InstanceProfile) clusterProfileName() string {
	return fmt.Sprintf("%s-%s", ip.ClusterID, ProfileNameTemplate)
}

func (ip *InstanceProfile) clusterRoleName() string {
	return fmt.Sprintf("%s-%s", ip.ClusterID, RoleNameTemplate)
}

func (ip *InstanceProfile) CreateIfNotExists() (bool, error) {
	return false, fmt.Errorf("instance profiles cannot be reused")
}

func (ip *InstanceProfile) CreateOrFail() error {
	if _, err := ip.Clients.IAM.CreateInstanceProfile(&iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(ip.clusterProfileName()),
	}); err != nil {
		return microerror.MaskAny(err)
	} else {
		if _, err := ip.Clients.IAM.AddRoleToInstanceProfile(&iam.AddRoleToInstanceProfileInput{
			InstanceProfileName: aws.String(ip.clusterProfileName()),
			RoleName:            aws.String(ip.clusterRoleName()),
		}); err != nil {
			return microerror.MaskAny(err)
		}
	}

	if err := ip.Clients.IAM.WaitUntilInstanceProfileExists(&iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(ip.clusterProfileName()),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	ip.name = ip.clusterProfileName()

	return nil
}

func (ip *InstanceProfile) Delete() error {
	if _, err := ip.Clients.IAM.DeleteInstanceProfile(&iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(ip.clusterProfileName()),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (ip *InstanceProfile) Name() string {
	return ip.name
}
