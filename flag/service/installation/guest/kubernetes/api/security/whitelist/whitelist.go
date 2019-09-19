package whitelist

type Whitelist struct {
	Private Private
	Public  Public
}

type Private struct {
	Enabled    bool
	SubnetList string
}

type Public struct {
	Enabled    bool
	SubnetList string
}
