package s3objectv1

import "github.com/aws/aws-sdk-go/service/s3"

type S3ClientMock struct {
}

func (s *S3ClientMock) PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return nil, nil
}

func (s *S3ClientMock) DeleteObject(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	return nil, nil
}
