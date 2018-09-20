package k8sportforward

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
)

type tunnelConfig struct {
	Dialer httpstream.Dialer

	RemotePort int
}

type Tunnel struct {
	localPort int
	stopChan  chan struct{}
}

func newTunnel(config tunnelConfig) (*Tunnel, error) {
	p, err := getAvailablePort()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	t := &Tunnel{
		localPort: p,
		stopChan:  make(chan struct{}, 1),
	}

	out := ioutil.Discard
	ports := []string{fmt.Sprintf("%d:%d", t.localPort, config.RemotePort)}
	readyChan := make(chan struct{}, 1)

	// next line will prevent `ERROR: logging before flag.Parse:` errors from
	// glog (used by k8s' apimachinery package) see
	// https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	flag.CommandLine.Parse([]string{})
	pf, err := portforward.New(config.Dialer, ports, t.stopChan, readyChan, out, out)
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
		return t, nil
	}
}

// Close disconnects a tunnel connection. It always returns nil error to fulfil
// io.Closer interface.
func (t *Tunnel) Close() error {
	close(t.stopChan)
	return nil
}

func (t *Tunnel) LocalAddress() string {
	return fmt.Sprintf("127.0.0.1:%d", t.localPort)
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

	return port, nil
}

// setConfigDefaults is copied and adjusted from client-go core/v1.
// TODO what is different here? Why do we need it at all?
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
