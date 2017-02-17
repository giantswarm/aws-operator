package aws

type Aws struct {
	KMSKeyArn string `json:"kmsKeyArn"`
	Masters   []Node `json:"masters"`
	Region    string `json:"region"`
	Workers   []Node `json:"workers"`
}
