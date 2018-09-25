package s3object

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeys"

	"github.com/giantswarm/aws-operator/service/controller/v16/encrypter"
)

const (
	// Name is the identifier of the resource.
	Name = "s3objectv16"
)

// Config represents the configuration used to create a new cloudformation resource.
type Config struct {
	CertWatcher       legacy.Searcher
	Encrypter         encrypter.Interface
	Logger            micrologger.Logger
	RandomKeySearcher randomkeys.Interface
}

// Resource implements the cloudformation resource.
type Resource struct {
	certWatcher       legacy.Searcher
	encrypter         encrypter.Interface
	logger            micrologger.Logger
	randomKeySearcher randomkeys.Interface
}

// New creates a new configured cloudformation resource.
func New(config Config) (*Resource, error) {
	if config.CertWatcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CertWatcher must not be empty")
	}
	if config.Encrypter == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Encrypter must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.RandomKeySearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.RandomKeySearcher must not be empty")
	}

	r := &Resource{
		certWatcher:       config.CertWatcher,
		encrypter:         config.Encrypter,
		logger:            config.Logger,
		randomKeySearcher: config.RandomKeySearcher,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func toBucketObjectState(v interface{}) (map[string]BucketObjectState, error) {
	if v == nil {
		return nil, nil
	}

	bucketObjectState, ok := v.(map[string]BucketObjectState)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", bucketObjectState, v)
	}

	return bucketObjectState, nil
}

func toPutObjectInput(v interface{}) (s3.PutObjectInput, error) {
	if v == nil {
		return s3.PutObjectInput{}, nil
	}

	bucketObject, ok := v.(BucketObjectState)
	if !ok {
		return s3.PutObjectInput{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", bucketObject, v)
	}

	putObjectInput := s3.PutObjectInput{
		Key:           aws.String(bucketObject.Key),
		Body:          strings.NewReader(bucketObject.Body),
		Bucket:        aws.String(bucketObject.Bucket),
		ContentLength: aws.Int64(int64(len(bucketObject.Body))),
	}

	return putObjectInput, nil
}
