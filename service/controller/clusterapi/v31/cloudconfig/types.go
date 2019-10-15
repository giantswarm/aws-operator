package cloudconfig

type templateData struct {
	AWSRegion      string
	CalicoMTU      int
	EncrypterType  string
	RegistryDomain string
	VaultAddress   string
	EncryptionKey  string
}
