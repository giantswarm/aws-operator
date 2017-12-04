package s3bucketv1

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/resource/cloudformationv1/adapter"
)

const (
	// Name is the identifier of the resource.
	Name = "s3bucket"
)

// Config represents the configuration used to create a new s3bucket resource.
type Config struct {
	// Dependencies.
	Clients Clients
	Logger  micrologger.Logger

	// Settings.
	AwsConfig awsutil.Config
}

// DefaultConfig provides a default configuration to create a new s3bucket
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Clients: Clients{},
		Logger:  nil,

		// Settings.
		AwsConfig: awsutil.Config{},
	}
}

// Resource implements the s3bucket resource.
type Resource struct {
	// Dependencies.
	clients Clients
	logger  micrologger.Logger

	// Settings.
	awsConfig awsutil.Config
}

// New creates a new configured s3bucket resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	// Settings.
	var emptyAwsConfig awsutil.Config
	if config.AwsConfig == emptyAwsConfig {
		return nil, microerror.Maskf(invalidConfigError, "config.AwsConfig must not be empty")
	}

	newResource := &Resource{
		// Dependencies.
		clients: config.Clients,
		logger: config.Logger.With(
			"resource", Name,
		),

		// Settings.
		awsConfig: config.AwsConfig,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func (r *Resource) getAccountID() (string, error) {
	resp, err := r.clients.IAM.GetUser(&iam.GetUserInput{})
	if err != nil {
		return "", microerror.Mask(err)
	}
	userArn := *resp.User.Arn
	accountID := strings.Split(userArn, ":")[accountIDIndex]
	if err := adapter.ValidateAccountID(accountID); err != nil {
		return "", microerror.Mask(err)
	}

	return accountID, nil
}
