package accountid

import (
	"github.com/aws/aws-sdk-go/service/sts"
)

type STS interface {
	GetCallerIdentity(*sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error)
}
