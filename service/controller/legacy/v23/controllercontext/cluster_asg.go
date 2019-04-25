package controllercontext

type ClusterASG struct {
	DesiredCapacity int
	MaxSize         int
	MinSize         int
}

func (a ClusterASG) IsEmpty() bool {
	return a.DesiredCapacity == 0 && a.MaxSize == 0 && a.MinSize == 0
}
