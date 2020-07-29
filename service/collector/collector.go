package collector

const (
	GaugeValue float64 = 1 // nolint:staticcheck
	namespace          = "aws_operator"
)

const (
	tagCluster      = "giantswarm.io/cluster"
	tagInstallation = "giantswarm.io/installation"
	tagOrganization = "giantswarm.io/organization"
)

const (
	labelAccount      = "account"
	labelAccountID    = "account_id"
	labelCluster      = "cluster_id"
	labelName         = "name"
	labelInstallation = "installation"
	labelOrganization = "organization"
)
