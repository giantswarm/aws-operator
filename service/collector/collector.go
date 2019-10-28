package collector

const (
	GaugeValue float64 = 1
	namespace          = "aws_operator"
)

const (
	tagCluster      = "giantswarm.io/cluster"
	tagInstallation = "giantswarm.io/installation"
	tagName         = "Name"
	tagOrganization = "giantswarm.io/organization"
	tagStack        = "aws:cloudformation:stack-name"
)

const (
	labelAccount      = "account"
	labelAccountID    = "account_id"
	labelAccountType  = "account_type"
	labelCluster      = "cluster_id"
	labelName         = "name"
	labelInstallation = "installation"
	labelOrganization = "organization"
)
