package accountid

import (
	"github.com/aws/aws-sdk-go/service/sts"
)

type Interface interface {
	Lookup() (string, error)
}

type STS interface {
	GetCallerIdentity(*sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error)
}
