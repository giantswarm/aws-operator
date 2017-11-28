package cloudformation

import (
	"regexp"

	"github.com/giantswarm/microerror"
)

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
