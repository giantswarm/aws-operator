package create

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

const (
	RoleNameTemplate         = "EC2-K8S-Role"
	PolicyNameTemplate       = "EC2-K8S-Policy"
	ProfileNameTemplate      = "EC2-K8S-Role"
	AssumeRolePolicyDocument = `{
		"Version": "2012-10-17",
		"Statement": {
			"Effect": "Allow",
			"Principal": {
				"Service": "ec2.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		}
	}`
	PolicyDocumentTempl = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "kms:Decrypt",
				"Resource": %q
			}
		]
	}`
)

func createRole(awsSession *session.Session, kmsKeyArn, clusterID string) error {
	svc := iam.New(awsSession)

	policyDocument := fmt.Sprintf(PolicyDocumentTempl, kmsKeyArn)

	clusterRoleName := fmt.Sprintf("%s-%s", clusterID, RoleNameTemplate)

	if _, err := svc.CreateRole(&iam.CreateRoleInput{
		RoleName:                 aws.String(clusterRoleName),
		AssumeRolePolicyDocument: aws.String(AssumeRolePolicyDocument),
	}); err != nil {
		return err
	}

	if _, err := svc.PutRolePolicy(&iam.PutRolePolicyInput{
		PolicyName:     aws.String(fmt.Sprintf("%s-%s", clusterID, PolicyNameTemplate)),
		RoleName:       aws.String(clusterRoleName),
		PolicyDocument: aws.String(policyDocument),
	}); err != nil {
		return err
	}

	return nil
}

func createInstanceProfile(awsSession *session.Session, clusterID string) (string, error) {
	svc := iam.New(awsSession)

	clusterRoleName := fmt.Sprintf("%s-%s", clusterID, RoleNameTemplate)
	clusterProfileName := fmt.Sprintf("%s-%s", clusterID, ProfileNameTemplate)

	if _, err := svc.CreateInstanceProfile(&iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(clusterProfileName),
	}); err != nil {
		return "", err
	} else {
		if _, err := svc.AddRoleToInstanceProfile(&iam.AddRoleToInstanceProfileInput{
			InstanceProfileName: aws.String(clusterProfileName),
			RoleName:            aws.String(clusterRoleName),
		}); err != nil {
			return "", err
		}
	}

	if err := svc.WaitUntilInstanceProfileExists(&iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(clusterProfileName),
	}); err != nil {
		return "", err
	}

	return clusterProfileName, nil
}
