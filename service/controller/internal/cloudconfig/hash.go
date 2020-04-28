package cloudconfig

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"strings"

	ignition "github.com/giantswarm/k8scloudconfig/v6/ignition/v_2_2_0"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

// hashIgnition returns a hash value representing the given ignition.
func hashIgnition(encoded []byte) string {
	rawSum := sha512.Sum512(encoded)
	sum := rawSum[:]
	encodedSum := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(encodedSum, sum)
	return string(encodedSum)
}

func (c *CloudConfig) DecryptedHash(ctx context.Context, data []byte) (string, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var decrypted ignition.Config
	err = json.Unmarshal(data, &decrypted)
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

	return hashIgnition(decryptedData), nil
}
