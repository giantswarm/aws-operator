package cloudconfig

type templateData struct {
	AWSRegion           string
	EncrypterType       string
	VaultAddress        string
	EncryptionKey       string
	MasterENIAddresses     []string
	MasterENIGateways    []string
	MasterENISubnetSize string
	MasterID            int
}
