package unittest

import "github.com/giantswarm/randomkeys/v2"

func DefaultKeys() randomkeys.Cluster {
	return randomkeys.Cluster{
		APIServerEncryptionKey: randomkeys.RandomKey("api-server-encryption-key"),
	}
}
