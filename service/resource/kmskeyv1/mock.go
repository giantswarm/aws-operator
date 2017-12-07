package kmskeyv1

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
)

type KMSClientMock struct {
	keyID     string
	aRN       string
	isError   bool
	clusterID string
}

func (k *KMSClientMock) CreateKey(input *kms.CreateKeyInput) (*kms.CreateKeyOutput, error) {
	return &kms.CreateKeyOutput{
		KeyMetadata: &kms.KeyMetadata{
			Arn:   aws.String("myarn"),
			KeyId: aws.String("mykeyid"),
		},
	}, nil
}

func (k *KMSClientMock) CreateAlias(input *kms.CreateAliasInput) (*kms.CreateAliasOutput, error) {
	if *input.AliasName != fmt.Sprintf("alias/%s", k.clusterID) {
		return nil, fmt.Errorf("unexpected alias, %v", input.AliasName)
	}

	if *input.TargetKeyId != "myarn" {
		return nil, fmt.Errorf("unexpected targetKeyID, %v", input.TargetKeyId)
	}

	return &kms.CreateAliasOutput{}, nil
}

func (k *KMSClientMock) DeleteAlias(input *kms.DeleteAliasInput) (*kms.DeleteAliasOutput, error) {
	return nil, nil
}

func (k *KMSClientMock) EnableKeyRotation(input *kms.EnableKeyRotationInput) (*kms.EnableKeyRotationOutput, error) {
	if *input.KeyId != "mykeyid" {
		return nil, fmt.Errorf("unexpected keyid, %v", input.KeyId)
	}

	return &kms.EnableKeyRotationOutput{}, nil
}

func (k *KMSClientMock) DescribeKey(input *kms.DescribeKeyInput) (*kms.DescribeKeyOutput, error) {
	if k.isError {
		return nil, fmt.Errorf("kms client failure")
	}
	output := &kms.DescribeKeyOutput{
		KeyMetadata: &kms.KeyMetadata{
			Arn:   aws.String(k.aRN),
			KeyId: aws.String(k.keyID),
		},
	}
	return output, nil
}
