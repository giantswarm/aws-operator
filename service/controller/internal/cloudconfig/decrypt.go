package cloudconfig

import (
	"context"
	"encoding/json"
	"fmt"
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
			content := file.Contents.Source
			ciphertext := strings.TrimPrefix(content, "data:text/plain;charset=utf-8;base64,")
			plaintext, err := c.encrypter.Decrypt(ctx, encryptionKey, ciphertext)
			if err != nil {
				return "", microerror.Mask(err)
			}
			content = fmt.Sprintf("data:text/plain;charset=utf-8;base64,%s", plaintext)
			decrypted.Storage.Files[i].Contents.Source = content
		}
	}

	decryptedData, err := json.Marshal(decrypted)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return string(decryptedData), nil
}
