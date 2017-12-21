package s3objectv2

import (
	"fmt"
	"io"
	"strings"

	"github.com/giantswarm/randomkeytpr"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certificatetpr"
)

// nopCloser is required to implement the ReadCloser interface required by
// the Body field in S3's GetObjectOutput
type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

type S3ClientMock struct {
	isError bool
	body    string
}

func (s *S3ClientMock) PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return nil, nil
}

func (s *S3ClientMock) DeleteObject(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	return nil, nil
}

func (s *S3ClientMock) GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	if s.isError {
		return nil, fmt.Errorf("error!!")
	}

	output := &s3.GetObjectOutput{
		Body: nopCloser{strings.NewReader(s.body)},
	}

	return output, nil
}

type CloudConfigMock struct {
	template string
}

func (c *CloudConfigMock) NewMasterTemplate(customObject v1alpha1.AWSConfig, certs certificatetpr.CompactTLSAssets, randomKeys randomkeytpr.CompactRandomKeyAssets) (string, error) {
	return c.template, nil
}

func (c *CloudConfigMock) NewWorkerTemplate(customObject v1alpha1.AWSConfig, certs certificatetpr.CompactTLSAssets) (string, error) {
	return c.template, nil
}

type KMSClientMock struct{}

func (k *KMSClientMock) Encrypt(input *kms.EncryptInput) (*kms.EncryptOutput, error) {
	return &kms.EncryptOutput{}, nil
}
