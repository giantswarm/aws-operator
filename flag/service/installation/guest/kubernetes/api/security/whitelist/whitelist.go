package whitelist

type Whitelist struct {
	Private Private
	Public  Public
}

type Private struct {
	Enabled    string
	SubnetList string
}

type Public struct {
	Enabled    string
	SubnetList string
}
