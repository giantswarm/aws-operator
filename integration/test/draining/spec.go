// +build k8srequired

package draining

type E2EAppResponse struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}
