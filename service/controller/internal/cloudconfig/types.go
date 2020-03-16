package cloudconfig

type templateData struct {
	AWSRegion      string
	EncryptionKey  string
	EncrypterType  string
	IsChinaRegion  bool
	RegistryDomain string
	VaultAddress   string
}
