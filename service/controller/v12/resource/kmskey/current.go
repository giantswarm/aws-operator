package kmskey

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/microerror"

	servicecontext "github.com/giantswarm/aws-operator/service/controller/v12/context"
	"github.com/giantswarm/aws-operator/service/controller/v12/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	currentState := KMSKeyState{}

	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return currentState, err
	}

	clusterID := key.ClusterID(customObject)
	alias := toAlias(clusterID)
	input := &kms.DescribeKeyInput{
		KeyId: aws.String(alias),
	}

	sc, err := servicecontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	output, err := sc.AWSClient.KMS.DescribeKey(input)
	if IsKeyNotFound(err) {
		// Fall through.
		return nil, nil
	}
	if err != nil {
		return currentState, microerror.Mask(err)
	}

	currentState.KeyID = *output.KeyMetadata.KeyId
	currentState.KeyAlias = alias

	return currentState, nil
}
