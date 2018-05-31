package s3object

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
)

func compactor(data []byte) (string, error) {
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if _, err := gzw.Write(data); err != nil {
		return "", err
	}
	if err := gzw.Close(); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func encryptor(svc KMSClient, kmsKeyARN string, data []byte) ([]byte, error) {
	/*
		encryptInput := kms.EncryptInput{
			KeyId:     aws.String(kmsKeyARN),
			Plaintext: data,
		}

		var encryptOutput *kms.EncryptOutput
		var err error
		if encryptOutput, err = svc.Encrypt(&encryptInput); err != nil {
			return []byte{}, err
		}
		return encryptOutput.CiphertextBlob, nil
	*/

	return data, nil
}
