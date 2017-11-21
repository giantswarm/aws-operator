// +build k8srequired

package integration

import (
	"github.com/giantswarm/microerror"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func getK8sClient() (kubernetes.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", "/workdir/.shipyard/config")
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return cs, nil
}
