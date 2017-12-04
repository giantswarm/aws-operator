package s3objectv1

import (
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
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

type AwsServiceMock struct {
	accountID string
	isError   bool
}

func (a AwsServiceMock) GetAccountID() (string, error) {
	if a.isError {
		return "", fmt.Errorf("error!!")
	}

	return a.accountID, nil
}
