package s3objectv1

import (
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/awstpr"
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

type CertWatcherMock struct {
	certs certificatetpr.AssetsBundle
}

func (c *CertWatcherMock) SearchCerts(string) (certificatetpr.AssetsBundle, error) {
	return c.certs, nil
}

type CloudConfigMock struct {
	template string
}

func (c *CloudConfigMock) NewWorkerTemplate(customObject awstpr.CustomObject, certs certificatetpr.CompactTLSAssets) (string, error) {
	return c.template, nil
}

type KMSClientMock struct{}

func (k *KMSClientMock) Encrypt(input *kms.EncryptInput) (*kms.EncryptOutput, error) {
	return &kms.EncryptOutput{}, nil
}
