package adapter

import (
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/giantswarm/microerror"
)

func AccountID(clients Clients) (string, error) {
	resp, err := clients.IAM.GetUser(&iam.GetUserInput{})
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
