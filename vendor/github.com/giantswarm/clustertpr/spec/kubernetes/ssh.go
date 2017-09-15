package kubernetes

import "github.com/giantswarm/clustertpr/spec/kubernetes/ssh"

type SSH struct {
	// UserList is a list of SSH  accounts with public key being added to each Kubernetes
	// node.
	UserList []ssh.User `json:"userList" yaml:"userList"`
}
