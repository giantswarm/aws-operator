package aws

type Interface interface {
	GetAccountID() (string, error)
	GetKeyArn(clusterID string) (string, error)
}
