package create

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
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
			},
			{
				"Effect": "Allow",
				"Action": [
					"s3:GetBucketLocation",
					"s3:ListAllMyBuckets"
				],
				"Resource": "*"
			},
			{
				"Effect": "Allow",
				"Action": [
					"s3:ListBucket"
				],
				"Resource": "arn:aws:s3:::%s"
			},
			{
				"Effect": "Allow",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::%s/%s/*"
			}
		]
	}`
)

func createRole(svc *iam.IAM, kmsKeyArn, s3Bucket, clusterID string) error {
	// TODO switch to using a file and Go templates
	policyDocument := fmt.Sprintf(PolicyDocumentTempl, kmsKeyArn, s3Bucket, s3Bucket, clusterID)

	clusterRoleName := fmt.Sprintf("%s-%s", clusterID, RoleNameTemplate)

	if _, err := svc.CreateRole(&iam.CreateRoleInput{
		RoleName:                 aws.String(clusterRoleName),
		AssumeRolePolicyDocument: aws.String(AssumeRolePolicyDocument),
	}); err != nil {
		return err
	}

	clusterPolicyName := fmt.Sprintf("%s-%s", clusterID, PolicyNameTemplate)

	if _, err := svc.PutRolePolicy(&iam.PutRolePolicyInput{
		PolicyName:     aws.String(clusterPolicyName),
		RoleName:       aws.String(clusterRoleName),
		PolicyDocument: aws.String(policyDocument),
	}); err != nil {
		return err
	}

	return nil
}

func deleteRole(svc *iam.IAM, clusterID string) error {
	clusterRoleName := fmt.Sprintf("%s-%s", clusterID, RoleNameTemplate)

	if _, err := svc.DeleteRole(&iam.DeleteRoleInput{
		RoleName: aws.String(clusterRoleName),
	}); err != nil {
		return err
	}

	return nil
}

func createInstanceProfile(svc *iam.IAM, clusterID string) (string, error) {
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

func deleteInstanceProfile(svc *iam.IAM, clusterID string) error {
	clusterProfileName := fmt.Sprintf("%s-%s", clusterID, ProfileNameTemplate)

	if _, err := svc.DeleteInstanceProfile(&iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(clusterProfileName),
	}); err != nil {
		return err
	}

	return nil
}

func removeRoleFromInstanceProfile(svc *iam.IAM, clusterID string) error {
	clusterRoleName := fmt.Sprintf("%s-%s", clusterID, RoleNameTemplate)
	clusterProfileName := fmt.Sprintf("%s-%s", clusterID, ProfileNameTemplate)

	if _, err := svc.RemoveRoleFromInstanceProfile(&iam.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: aws.String(clusterProfileName),
		RoleName:            aws.String(clusterRoleName),
	}); err != nil {
		return err
	}

	return nil
}

func deletePolicy(svc *iam.IAM, clusterID string) error {
	clusterRoleName := fmt.Sprintf("%s-%s", clusterID, RoleNameTemplate)
	clusterPolicyName := fmt.Sprintf("%s-%s", clusterID, PolicyNameTemplate)

	if _, err := svc.DeleteRolePolicy(&iam.DeleteRolePolicyInput{
		RoleName:   aws.String(clusterRoleName),
		PolicyName: aws.String(clusterPolicyName),
	}); err != nil {
		return err
	}

	return nil
}

func deletePolicyResources(svc *iam.IAM, clusterID string) error {
	if err := removeRoleFromInstanceProfile(svc, clusterID); err != nil {
		return err
	}

	if err := deleteInstanceProfile(svc, clusterID); err != nil {
		return err
	}

	if err := deletePolicy(svc, clusterID); err != nil {
		return err
	}

	if err := deleteRole(svc, clusterID); err != nil {
		return err
	}

	return nil
}
