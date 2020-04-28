package cloudconfig

import (
	"crypto/sha512"
	"encoding/hex"
)

// hashIgnition returns a hash value representing the given ignition.
func hashIgnition(encoded []byte) (string, error) {
	rawSum := sha512.Sum512(encoded)
	sum := rawSum[:]
	encodedSum := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(encodedSum, sum)
	return string(encodedSum), nil
}
