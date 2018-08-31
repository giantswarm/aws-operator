package k8sportforward

import (
	"net/http"

	"github.com/giantswarm/microerror"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport/spdy"
)

type ForwarderConfig struct {
	RestConfig *rest.Config
}

type Forwarder struct {
	restConfig *rest.Config

	k8sClient rest.Interface
}

func NewForwarder(config ForwarderConfig) (*Forwarder, error) {
	if config.RestConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.RestConfig must not be empty", config)
	}

	setConfigDefaults(config.RestConfig)

	k8sClient, err := rest.RESTClientFor(config.RestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	f := &Forwarder{
		restConfig: config.RestConfig,

		k8sClient: k8sClient,
	}

	return f, nil
}

// ForwardPort opens a tunnel to a kubernetes pod.
func (f *Forwarder) ForwardPort(namespace string, podName string, remotePort int) (*Tunnel, error) {
	transport, upgrader, err := spdy.RoundTripperFor(f.restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	config := tunnelConfig{
		Dialer: spdy.NewDialer(
			upgrader,
			&http.Client{
				Transport: transport,
			},
			"POST",
			// Build a url to the portforward endpoint.
			// Example: http://localhost:8080/api/v1/namespaces/helm/pods/tiller-deploy-9itlq/portforward
			f.k8sClient.Post().Resource("pods").Namespace(namespace).Name(podName).SubResource("portforward").URL(),
		),

		RemotePort: remotePort,
	}

	tunnel, err := newTunnel(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tunnel, nil
}
