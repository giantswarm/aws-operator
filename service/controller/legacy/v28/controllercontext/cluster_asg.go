package controllercontext

type ContextStatusTenantClusterTCCPASG struct {
	DesiredCapacity int
	MaxSize         int
	MinSize         int
	Name            string
}

func (a ContextStatusTenantClusterTCCPASG) IsEmpty() bool {
	return a.DesiredCapacity == 0 && a.MaxSize == 0 && a.MinSize == 0
}
