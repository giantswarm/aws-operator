package s3object

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/v13/service/internal/cloudconfig"
	"github.com/giantswarm/aws-operator/v13/service/internal/encrypter"
)

const (
	// Name is the identifier of the resource.
	Name = "s3object"
)

// Config represents the configuration used to create a new cloudformation resource.
type Config struct {
	CloudConfig cloudconfig.Interface
	Encrypter   encrypter.Interface
	Logger      micrologger.Logger
}

// Resource implements the CRUD resource interface of operatorkit to manage S3
// objects containing rendered Cloud Config templates. The current
// implementation potentially causes some amount of S3 Traffic which might be
// neglectable from a customer costs point of view. In order to improve the
// implementation and its performance as well as produced costs we would need to
// refactor the resource and rewrite it basically completely using the simple
// resource interface only using EnsureCreated and EnsureDeleted. That way we
// could compute the E-Tag for the request to fetch the S3 Object and reduce
// traffic as we would not fetch the whole body of the object when it does
// effectively not change most of the time.
type Resource struct {
	cloudConfig cloudconfig.Interface
	encrypter   encrypter.Interface
	logger      micrologger.Logger
}

// New creates a new configured cloudformation resource.
func New(config Config) (*Resource, error) {
	if config.CloudConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CloudConfig must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		cloudConfig: config.CloudConfig,
		encrypter:   config.Encrypter,
		logger:      config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func toS3Objects(v interface{}) ([]*s3.PutObjectInput, error) {
	if v == nil {
		return nil, nil
	}

	t, ok := v.([]*s3.PutObjectInput)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*s3.PutObjectInput{}, v)
	}

	return t, nil
}
