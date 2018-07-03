package provider

type Interface interface {
	RebootMaster() error
	ReplaceMaster() error
}
