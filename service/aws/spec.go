package aws

import (
	"github.com/aws/aws-sdk-go/service/iam"
)

type Clients struct {
	IAM IAMClient
}

// IAMClient describes the methods required to be implemented by a IAM AWS client.
type IAMClient interface {
	GetUser(*iam.GetUserInput) (*iam.GetUserOutput, error)
}
