package aws

type AWS struct {
	KMSKeyArn string `json:"kmsKeyArn" yaml:"kmsKeyArn"`
	Masters   []Node `json:"masters" yaml:"masters"`
	Region    string `json:"region" yaml:"region"`
	AZ        string `json:"az" yaml:"az"`
	Workers   []Node `json:"workers" yaml:"workers"`
}
