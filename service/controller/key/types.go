package key

type AMIInfo struct {
	Name string `json:"name"`
	PV   string `json:"pv"`
	HVM  string `json:"hvm"`
}

type AMIInfoList struct {
	AMIs []AMIInfo `json:"amis"`
}
