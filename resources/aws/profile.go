package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/giantswarm/microerror"
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
		return microerror.Mask(err)
	} else {
		if _, err := ip.Clients.IAM.AddRoleToInstanceProfile(&iam.AddRoleToInstanceProfileInput{
			InstanceProfileName: aws.String(ip.clusterProfileName()),
			RoleName:            aws.String(ip.clusterRoleName()),
		}); err != nil {
			return microerror.Mask(err)
		}
	}

	if err := ip.Clients.IAM.WaitUntilInstanceProfileExists(&iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(ip.clusterProfileName()),
	}); err != nil {
		return microerror.Mask(err)
	}

	ip.name = ip.clusterProfileName()

	return nil
}

func (ip *InstanceProfile) Delete() error {
	if _, err := ip.Clients.IAM.DeleteInstanceProfile(&iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(ip.clusterProfileName()),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (ip InstanceProfile) GetName() string {
	return ip.name
}
