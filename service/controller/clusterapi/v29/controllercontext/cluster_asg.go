package controllercontext

func (a ContextStatusTenantClusterTCCPASG) IsEmpty() bool {
	return a.DesiredCapacity == 0 && a.MaxSize == 0 && a.MinSize == 0
}
