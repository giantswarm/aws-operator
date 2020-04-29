package cloudconfig

import (
	"context"
	"encoding/json"
	"strings"

	ignition "github.com/giantswarm/k8scloudconfig/v6/ignition/v_2_2_0"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

func (c *CloudConfig) DecryptTemplate(ctx context.Context, data string) (string, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var decrypted ignition.Config
	err = json.Unmarshal([]byte(data), &decrypted)
	if err != nil {
		return "", microerror.Mask(err)
	}

	encryptionKey := cc.Status.TenantCluster.Encryption.Key
	for i, file := range decrypted.Storage.Files {
		if strings.HasSuffix(file.Path, ".enc") {
			decrypted.Storage.Files[i].Contents.Source, err = c.encrypter.Decrypt(ctx, encryptionKey, file.Contents.Source)
			if err != nil {
				return "", microerror.Mask(err)
			}
		}
	}

	decryptedData, err := json.Marshal(decrypted)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return string(decryptedData), nil
}
