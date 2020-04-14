package controllercontext

func (a ContextStatusTenantClusterASG) IsEmpty() bool {
	return a.DesiredCapacity == 0 && a.MaxSize == 0 && a.MinSize == 0
}
