package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/kms"
	microerror "github.com/giantswarm/microkit/error"
)

type KMSKey struct {
	arn string
	AWSEntity
}

func (kk *KMSKey) CreateIfNotExists() (bool, error) {
	return false, fmt.Errorf("KMS keys cannot be reused")
}

func (kk *KMSKey) CreateOrFail() error {
	// TODO we should give it a name
	key, err := kk.Clients.KMS.CreateKey(&kms.CreateKeyInput{})
	if err != nil {
		return microerror.MaskAny(err)
	}
	kk.arn = *key.KeyMetadata.Arn
	return nil
}

func (kms *KMSKey) Delete() error {
	return microerror.MaskAny(notImplementedMethodError)
}

func (kms *KMSKey) Arn() string {
	return kms.arn
}
