package cloudconfig

func encryptor(kmsClient KMSClient, kmsKeyARN string, data []byte) ([]byte, error) {
	/*
		encryptInput := &kms.EncryptInput{
			KeyId:     aws.String(kmsKeyARN),
			Plaintext: data,
		}

		encryptOutput, err := kmsClient.Encrypt(encryptInput)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		return encryptOutput.CiphertextBlob, nil
	*/

	return data, nil
}
