package accountid

import (
	"regexp"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	// accountIDIndex represents the index in which we can find the account ID in
	// the user ARN, splitting by colon.
	accountIDIndex = 4
	// accountIDLength is the length of the parsed string which we assume is a
	// valid account ID.
	accountIDLength = 12
)

type Config struct {
	Logger micrologger.Logger
	STS    STS
}

type AccountID struct {
	logger micrologger.Logger
	sts    STS

	accountID string
	mutex     sync.Mutex
}

func New(config Config) (*AccountID, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.STS == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.STS must not be empty", config)
	}

	a := &AccountID{
		logger: config.Logger,
		sts:    config.STS,

		accountID: "",
		mutex:     sync.Mutex{},
	}

	return a, nil
}

func (a *AccountID) Lookup() (string, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.accountID != "" {
		return a.accountID, nil
	}

	accountID, err := a.lookup()
	if err != nil {
		return "", microerror.Mask(err)
	}
	a.accountID = accountID

	return accountID, nil
}

func (a *AccountID) lookup() (string, error) {
	var arn string
	{
		i := &sts.GetCallerIdentityInput{}

		o, err := a.sts.GetCallerIdentity(i)
		if err != nil {
			return "", microerror.Mask(err)
		}

		arn = *o.Arn
	}

	var id string
	{
		id = strings.Split(arn, ":")[accountIDIndex]

		err := validateAccountID(id)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	return id, nil
}

func validateAccountID(id string) error {
	r, _ := regexp.Compile("^[0-9]*$")

	switch {
	case id == "":
		return microerror.Maskf(invalidAccountIDError, "account ID must not be empty")
	case len(id) != accountIDLength:
		return microerror.Maskf(invalidAccountIDError, "account ID must have length %d", accountIDLength)
	case !r.MatchString(id):
		return microerror.Maskf(invalidAccountIDError, "account ID must match %#q", r.String())
	}

	return nil
}
