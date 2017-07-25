package kubernetes

import (
	"net"
)

type DNS struct {
	// IP is the kube DNS cluster IP, e.g. 172.31.0.10.
	IP net.IP `json:"ip" yaml:"ip"`
}
