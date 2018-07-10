package provider

type GuestClusterConfig struct {
	Name string
}

type Interface interface {
	RequestGuestClusterCreation(clusterName string) error
	RequestGuestClusterDeletion(clusterName string)
}
