package etcd

import (
	"net"
)

type Etcd struct {
	Domain string `json:"domain" yaml:"domain"`
	IP     net.IP `json:"ip" yaml:"ip"`
	Port   int    `json:"port" yaml:"port"`
	Prefix string `json:"prefix" yaml:"prefix"`
}
