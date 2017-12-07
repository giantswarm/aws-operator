package kmskeyv1

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
)

type KMSClientMock struct {
	keyID   string
	aRN     string
	isError bool
}

func (k *KMSClientMock) CreateKey(input *kms.CreateKeyInput) (*kms.CreateKeyOutput, error) {
	return nil, nil
}

func (k *KMSClientMock) CreateAlias(input *kms.CreateAliasInput) (*kms.CreateAliasOutput, error) {
	return nil, nil
}

func (k *KMSClientMock) DeleteAlias(input *kms.DeleteAliasInput) (*kms.DeleteAliasOutput, error) {
	return nil, nil
}

func (k *KMSClientMock) EnableKeyRotation(input *kms.EnableKeyRotationInput) (*kms.EnableKeyRotationOutput, error) {
	return nil, nil
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
