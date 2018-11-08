package release

type ChartInfo struct {
	isChannel bool
	name      string
	version   string
}

func NewChartInfo(name string, gitSHA string) ChartInfo {
	return ChartInfo{
		isChannel: false,
		name:      name,
		version:   "1.0.0-" + gitSHA,
	}
}

func NewStableChartInfo(name string) ChartInfo {
	return ChartInfo{
		isChannel: true,
		name:      name,
		version:   "stable",
	}
}

// Version is an **alias** for ChartInfo. See https://golang.org/doc/go1.9#language.
type Version = ChartInfo

// NewStableVersion is a compatibility method used for transition period only.
// Use NewStableChartInfo instead. For the transition period both versions of
// the code will have the same semantics.
//
// 	r.Install(ctx, "apiextensions-aws-config-e2e", NewStableChartInfo("apiextensions-aws-config-e2e-chart"), values)
//	r.Install(ctx, "apiextensions-aws-config-e2e", NewStableVersion(), values))
//
// The new ChartInfo struct was introduced to solve the problem of installing
// multiple releases of the same chart. E.g.
//
// 	r.Install(ctx, "release-1", NewStableChartInfo("apiextensions-aws-config-e2e-chart"), values)
// 	r.Install(ctx, "release-2", NewStableChartInfo("apiextensions-aws-config-e2e-chart"), values)
//
func NewStableVersion() Version {
	return ChartInfo{
		version:   "stable",
		isChannel: true,
	}
}

// NewVersion is a compatibility method used for transition period only. Use
// NewChartInfo instead. For the transition period both versions of the code
// will have the same semantics.
//
// 	r.Install(ctx, "apiextensions-aws-config-e2e", NewChartInfo("apiextensions-aws-config-e2e-chart", "afae404f496389dd955e70dfc78e898aa6726265"), values)
//	r.Install(ctx, "apiextensions-aws-config-e2e", NewVersion("afae404f496389dd955e70dfc78e898aa6726265"), values))
//
// The new ChartInfo struct was introduced to solve the problem of installing
// multiple releases of the same chart. E.g.
//
// 	r.Install(ctx, "release-1", NewChartInfo("apiextensions-aws-config-e2e-chart", "afae404f496389dd955e70dfc78e898aa6726265"), values)
// 	r.Install(ctx, "release-2", NewChartInfo("apiextensions-aws-config-e2e-chart", "afae404f496389dd955e70dfc78e898aa6726265"), values)
//
func NewVersion(gitSHA string) Version {
	return ChartInfo{
		version:   "1.0.0-" + gitSHA,
		isChannel: false,
	}
}
