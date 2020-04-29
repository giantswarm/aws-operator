package cloudconfig

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"

	ignition "github.com/giantswarm/k8scloudconfig/v6/ignition/v_2_2_0"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

// hashIgnition returns a hash value representing the given ignition.
func hashIgnition(encoded string, replacements map[string]string, logger micrologger.Logger) (string, error) {
	var config ignition.Config
	err := json.Unmarshal([]byte(encoded), &config)
	if err != nil {
		return "", microerror.Mask(err)
	}
	for i, file := range config.Storage.Files {
		if replacement, ok := replacements[file.Path]; ok {
			logger.LogCtx(context.Background(), "level", "debug", "message", fmt.Sprintf("replacing %s: %s %s", file.Path, config.Storage.Files[i].Contents.Source, replacement))
			config.Storage.Files[i].Contents.Source = replacement
		}
	}
	reencoded, err := json.Marshal(config)
	logger.LogCtx(context.Background(), "level", "debug", "message", fmt.Sprintf("template with replacements %s", string(reencoded)))
	if err != nil {
		return "", microerror.Mask(err)
	}
	rawSum := sha512.Sum512(reencoded)
	sum := rawSum[:]
	encodedSum := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(encodedSum, sum)
	return string(encodedSum), nil
}
