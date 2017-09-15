package ssh

type User struct {
	Name      string `json:"name" yaml:"name"`
	PublicKey string `json:"publicKey" yaml:"publicKey"`
}
