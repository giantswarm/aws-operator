package k8sportforward

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type Config struct {
	RestConfig *rest.Config
}

type Forwarder struct {
	k8sClient  rest.Interface
	restConfig *rest.Config
}

func New(config Config) (*Forwarder, error) {
	if config.RestConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.RestConfig must not be empty")
	}

	setConfigDefaults(config.RestConfig)

	k8sClient, err := rest.RESTClientFor(config.RestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &Forwarder{
		k8sClient:  k8sClient,
		restConfig: config.RestConfig,
	}, nil
}

// ForwardPort opens a tunnel to a kubernetes pod.
func (f *Forwarder) ForwardPort(config TunnelConfig) (*Tunnel, error) {
	// Build a url to the portforward endpoint.
	// Example: http://localhost:8080/api/v1/namespaces/helm/pods/tiller-deploy-9itlq/portforward
	u := f.k8sClient.Post().
		Resource("pods").
		Namespace(config.Namespace).
		Name(config.PodName).
		SubResource("portforward").URL()

	transport, upgrader, err := spdy.RoundTripperFor(f.restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", u)

	tunnel, err := newTunnel(dialer, config)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return tunnel, err
}

type TunnelConfig struct {
	Remote    int
	Namespace string
	PodName   string
}

type Tunnel struct {
	TunnelConfig
	Local int

	stopChan chan struct{}
}

// Close disconnects a tunnel connection. It always returns nil error to fulfil
// io.Closer interface.
func (t *Tunnel) Close() error {
	close(t.stopChan)
	return nil
}

func newTunnel(dialer httpstream.Dialer, config TunnelConfig) (*Tunnel, error) {
	local, err := getAvailablePort()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	tunnel := &Tunnel{
		TunnelConfig: config,
		Local:        local,

		stopChan: make(chan struct{}, 1),
	}

	out := ioutil.Discard
	ports := []string{fmt.Sprintf("%d:%d", tunnel.Local, tunnel.Remote)}
	readyChan := make(chan struct{}, 1)

	// next line will prevent `ERROR: logging before flag.Parse:` errors from
	// glog (used by k8s' apimachinery package) see
	// https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	flag.CommandLine.Parse([]string{})
	pf, err := portforward.New(dialer, ports, tunnel.stopChan, readyChan, out, out)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errChan := make(chan error)
	go func() {
		select {
		case errChan <- pf.ForwardPorts():
		case <-ctx.Done():
		}
	}()

	select {
	case err = <-errChan:
		return nil, microerror.Mask(err)
	case <-pf.Ready:
		return tunnel, nil
	}
}

func getAvailablePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, microerror.Mask(err)
	}
	defer l.Close()

	_, p, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return 0, microerror.Mask(err)
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return 0, microerror.Mask(err)
	}
	return port, microerror.Mask(err)
}

// setConfigDefaults is copied and adjusted from client-go core/v1.
func setConfigDefaults(config *rest.Config) error {
	if config.GroupVersion == nil {
		config.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	}
	if config.APIPath == "" {
		config.APIPath = "/api"
	}
	if config.NegotiatedSerializer == nil {
		s := runtime.NewScheme()
		c := serializer.NewCodecFactory(s)
		config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: c}
	}
	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}
