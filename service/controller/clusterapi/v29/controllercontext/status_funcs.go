package controllercontext

func (a ContextStatusTenantClusterTCNPASG) IsEmpty() bool {
	return a.DesiredCapacity == 0 && a.MaxSize == 0 && a.MinSize == 0
}
