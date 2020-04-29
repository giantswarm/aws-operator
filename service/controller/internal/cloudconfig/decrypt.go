package cloudconfig

import (
	"context"
	"encoding/base64"
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
			ciphertextEncoded := strings.TrimPrefix(content, "data:text/plain;charset=utf-8;base64,")
			ciphertext, err := base64.StdEncoding.DecodeString(ciphertextEncoded)
			if err != nil {
				return "", microerror.Mask(err)
			}
			plaintext, err := c.encrypter.Decrypt(ctx, encryptionKey, string(ciphertext))
			if err != nil {
				return "", microerror.Mask(err)
			}
			encoded := base64.StdEncoding.EncodeToString([]byte(plaintext))
			content = fmt.Sprintf("data:text/plain;charset=utf-8;base64,%s", encoded)
			decrypted.Storage.Files[i].Contents.Source = content
		}
	}

	decryptedData, err := json.Marshal(decrypted)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return string(decryptedData), nil
}
