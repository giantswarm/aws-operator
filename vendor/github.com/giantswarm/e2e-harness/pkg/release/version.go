package release

type Version struct {
	name      string
	isChannel bool
}

func NewStableVersion() Version {
	return Version{
		name:      "stable",
		isChannel: true,
	}
}

func NewVersion(gitSHA string) Version {
	return Version{
		name:      "1.0.0-" + gitSHA,
		isChannel: false,
	}
}

func (v Version) String() string {
	return v.name
}
