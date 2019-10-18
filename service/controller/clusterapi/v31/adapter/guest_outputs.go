package adapter

type GuestOutputsAdapter struct {
	IngressInsecureTargetGroupResourceName string
	IngressSecureTargetGroupResourceName   string
	Master                                 GuestOutputsAdapterMaster
	OperatorVersion                        string
	Route53Enabled                         bool
}

func (a *GuestOutputsAdapter) Adapt(config Config) error {
	a.IngressInsecureTargetGroupResourceName = ingressELBInsecureTargetGroupResourceName
	a.IngressSecureTargetGroupResourceName = ingressELBSecureTargetGroupResourceName

	a.Master.DockerVolume.ResourceName = config.StackState.DockerVolumeResourceName
	a.Master.ImageID = config.StackState.MasterImageID
	a.Master.Instance.ResourceName = config.StackState.MasterInstanceResourceName
	a.Master.Instance.Type = config.StackState.MasterInstanceType

	a.OperatorVersion = config.StackState.OperatorVersion

	a.Route53Enabled = config.Route53Enabled

	return nil
}

type GuestOutputsAdapterMaster struct {
	ImageID      string
	Instance     GuestOutputsAdapterMasterInstance
	DockerVolume GuestOutputsAdapterMasterDockerVolume
}

type GuestOutputsAdapterMasterInstance struct {
	ResourceName string
	Type         string
}

type GuestOutputsAdapterMasterDockerVolume struct {
	ResourceName string
}
