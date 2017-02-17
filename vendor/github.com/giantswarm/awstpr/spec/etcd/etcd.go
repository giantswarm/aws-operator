package etcd

type Etcd struct {
	Domain string `json:"domain"`
	IP     string `json:"ip"`
	Port   string `json:"port"`
	Prefix string `json:"prefix"`
}
