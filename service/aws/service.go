package aws

import (
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

// Config represents the configuration used to create a new aws service.
type Config struct {
	// Dependencies.
	Clients Clients
	Logger  micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new aws service
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Clients: Clients{},
		Logger:  nil,
	}
}

// Service implements the aws service.
type Service struct {
	// Dependencies.
	clients Clients
	logger  micrologger.Logger
}

// New creates a new configured aws service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Service{
		// Dependencies.
		clients: config.Clients,
		logger:  config.Logger,
	}

	return newService, nil
}

// GetAccountID gets the AWS Account ID.
func (s *Service) GetAccountID() (string, error) {
	resp, err := s.clients.IAM.GetUser(&iam.GetUserInput{})
	if err != nil {
		return "", microerror.Mask(err)
	}
	userArn := *resp.User.Arn
	accountID := strings.Split(userArn, ":")[accountIDIndex]
	if err := ValidateAccountID(accountID); err != nil {
		return "", microerror.Mask(err)
	}

	return accountID, nil
}

// ValidateAccountID validates the AWS Account ID.
func ValidateAccountID(accountID string) error {
	r, _ := regexp.Compile("^[0-9]*$")

	switch {
	case accountID == "":
		return microerror.Mask(emptyAmazonAccountIDError)
	case len(accountID) != accountIDLength:
		return microerror.Mask(wrongAmazonAccountIDLengthError)
	case !r.MatchString(accountID):
		return microerror.Mask(malformedAmazonAccountIDError)
	}

	return nil
}
