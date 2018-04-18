package cloudconfig

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"

	"github.com/giantswarm/microerror"
)

func compactor(data []byte) (string, error) {
	var err error

	buf := &bytes.Buffer{}
	gzw := gzip.NewWriter(buf)

	_, err = gzw.Write(data)
	if err != nil {
		return "", microerror.Mask(err)
	}

	err = gzw.Close()
	if err != nil {
		return "", microerror.Mask(err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
