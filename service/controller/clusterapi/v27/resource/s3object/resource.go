package s3object

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeys"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/cloudconfig"
)

const (
	// Name is the identifier of the resource.
	Name = "s3objectv27"
)

// Config represents the configuration used to create a new cloudformation resource.
type Config struct {
	CertsSearcher      certs.Interface
	CloudConfig        cloudconfig.Interface
	Logger             micrologger.Logger
	RandomKeysSearcher randomkeys.Interface
}

// Resource implements the cloudformation resource.
type Resource struct {
	certsSearcher      certs.Interface
	cloudConfig        cloudconfig.Interface
	logger             micrologger.Logger
	randomKeysSearcher randomkeys.Interface
}

// New creates a new configured cloudformation resource.
func New(config Config) (*Resource, error) {
	if config.CertsSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CertsSearcher must not be empty", config)
	}
	if config.CloudConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CloudConfig must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.RandomKeysSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.RandomKeySearcher must not be empty", config)
	}

	r := &Resource{
		certsSearcher:      config.CertsSearcher,
		cloudConfig:        config.CloudConfig,
		logger:             config.Logger,
		randomKeysSearcher: config.RandomKeysSearcher,
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
