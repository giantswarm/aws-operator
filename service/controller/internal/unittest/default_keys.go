package unittest

import "github.com/giantswarm/randomkeys"

func DefaultKeys() randomkeys.Cluster {
	return randomkeys.Cluster{
		APIServerEncryptionKey: randomkeys.RandomKey("api-server-encryption-key"),
	}
}
