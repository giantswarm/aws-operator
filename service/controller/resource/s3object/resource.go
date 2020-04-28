package s3object

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeys"

	"github.com/giantswarm/aws-operator/service/controller/internal/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	// Name is the identifier of the resource.
	Name = "s3object"
)

// Config represents the configuration used to create a new cloudformation resource.
type Config struct {
	CertsSearcher      certs.Interface
	CloudConfig        cloudconfig.Interface
	G8sClient          versioned.Interface
	LabelsFunc         func(key.LabelsGetter) string
	Logger             micrologger.Logger
	PathFunc           func(key.LabelsGetter) string
	RandomKeysSearcher randomkeys.Interface
	RegistryDomain     string
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
	certsSearcher      certs.Interface
	cloudConfig        cloudconfig.Interface
	g8sClient          versioned.Interface
	labelsFunc         func(key.LabelsGetter) string
	logger             micrologger.Logger
	pathFunc           func(key.LabelsGetter) string
	randomKeysSearcher randomkeys.Interface
	registryDomain     string
}

// New creates a new configured cloudformation resource.
func New(config Config) (*Resource, error) {
	if config.CertsSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CertsSearcher must not be empty", config)
	}
	if config.CloudConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CloudConfig must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.LabelsFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.LabelsFunc must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.PathFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.PathFunc must not be empty", config)
	}
	if config.RandomKeysSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.RandomKeySearcher must not be empty", config)
	}
	if config.RegistryDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RegistryDomain must not be empty", config)
	}

	r := &Resource{
		certsSearcher:      config.CertsSearcher,
		cloudConfig:        config.CloudConfig,
		g8sClient:          config.G8sClient,
		labelsFunc:         config.LabelsFunc,
		logger:             config.Logger,
		pathFunc:           config.PathFunc,
		randomKeysSearcher: config.RandomKeysSearcher,
		registryDomain:     config.RegistryDomain,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func toS3Object(v interface{}) (*s3.PutObjectInput, error) {
	if v == nil {
		return nil, nil
	}

	t, ok := v.(*s3.PutObjectInput)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &s3.PutObjectInput{}, v)
	}

	return t, nil
}

func toS3ObjectArray(v interface{}) ([]*s3.PutObjectInput, error) {
	if v == nil {
		return nil, nil
	}

	t, ok := v.([]*s3.PutObjectInput)
	if !ok {
		return []*s3.PutObjectInput{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*s3.PutObjectInput{{}}, v)
	}

	return t, nil
}
