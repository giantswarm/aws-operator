package collector

const (
	GaugeValue float64 = 1
	namespace  string  = "aws_operator"
)

const (
	tagCluster      = "giantswarm.io/cluster"
	tagName         = "Name"
	tagOrganization = "giantswarm.io/organization"
	tagStackName    = "aws:cloudformation:stack-name"
	tagStack        = "giantswarm.io/stack"
)

const (
	labelAccount      = "account"
	labelAccountID    = "account_id"
	labelCluster      = "cluster_id"
	labelName         = "name"
	labelInstallation = "installation"
	labelOrganization = "organization"
)
