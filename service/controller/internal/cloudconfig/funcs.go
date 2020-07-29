package cloudconfig

import (
	"crypto/sha1" // nolint:gosec
	"encoding/hex"
	"encoding/json"

	"github.com/giantswarm/k8scloudconfig/v6/pkg/ignition"
	"github.com/giantswarm/microerror"
)

// hashIgnition returns a hash value representing the given ignition.
func hashIgnition(encoded string, replacements map[string]string) (string, error) {
	var config ignition.Config
	err := json.Unmarshal([]byte(encoded), &config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	for i, file := range config.Storage.Files {
		if replacement, ok := replacements[file.Path]; ok {
			config.Storage.Files[i].Contents.Source = replacement
		}
	}
	reencoded, err := json.Marshal(config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	rawSum := sha1.Sum(reencoded) // nolint:gosec
	sum := rawSum[:]
	encodedSum := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(encodedSum, sum)
	return string(encodedSum), nil
}
