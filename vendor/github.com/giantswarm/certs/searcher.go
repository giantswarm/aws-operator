package certs

import (
	"fmt"
	"sync"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

const (
	// DefaultWatchTimeout is the time to wait on watches against the Kubernetes
	// API before giving up and throwing an error.
	DefaultWatchTimeout = 90 * time.Second
)

type Config struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	WatchTimeout time.Duration
}

type Searcher struct {
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	watchTimeout time.Duration
}

func NewSearcher(config Config) (*Searcher, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.WatchTimeout == 0 {
		config.WatchTimeout = DefaultWatchTimeout
	}

	s := &Searcher{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		watchTimeout: config.WatchTimeout,
	}

	return s, nil
}

func (s *Searcher) SearchAppOperator(clusterID string) (AppOperator, error) {
	var appOperator AppOperator

	certificates := []struct {
		TLS  *TLS
		Cert Cert
	}{
		{TLS: &appOperator.APIServer, Cert: AppOperatorAPICert},
	}

	g := &errgroup.Group{}
	m := sync.Mutex{}

	for _, certificate := range certificates {
		c := certificate

		g.Go(func() error {
			secret, err := s.search(c.TLS, clusterID, c.Cert)
			if err != nil {
				return microerror.Mask(err)
			}

			m.Lock()
			defer m.Unlock()
			err = fillTLSFromSecret(c.TLS, secret, clusterID, c.Cert)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		return AppOperator{}, microerror.Mask(err)
	}

	return appOperator, nil
}

func (s *Searcher) SearchCluster(clusterID string) (Cluster, error) {
	var cluster Cluster

	certificates := []struct {
		TLS      *TLS
		Cert     Cert
		optional bool
	}{
		{TLS: &cluster.APIServer, Cert: APICert},
		{TLS: &cluster.CalicoEtcdClient, Cert: CalicoEtcdClientCert, optional: true},
		{TLS: &cluster.EtcdServer, Cert: EtcdCert},
		{TLS: &cluster.ServiceAccount, Cert: ServiceAccountCert},
		{TLS: &cluster.Worker, Cert: WorkerCert},
	}

	g := &errgroup.Group{}
	m := sync.Mutex{}

	for _, certificate := range certificates {
		c := certificate

		g.Go(func() error {
			secret, err := s.search(c.TLS, clusterID, c.Cert)
			if err != nil {
				if c.optional {
					s.logger.Log("level", "warning", "message", err.Error())
					return nil
				} else {
					return microerror.Mask(err)
				}
			}

			m.Lock()
			defer m.Unlock()
			err = fillTLSFromSecret(c.TLS, secret, clusterID, c.Cert)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		return Cluster{}, microerror.Mask(err)
	}

	return cluster, nil
}

func (s *Searcher) SearchClusterOperator(clusterID string) (ClusterOperator, error) {
	var clusterOperator ClusterOperator

	certificates := []struct {
		TLS  *TLS
		Cert Cert
	}{
		{TLS: &clusterOperator.APIServer, Cert: ClusterOperatorAPICert},
	}

	g := &errgroup.Group{}
	m := sync.Mutex{}

	for _, certificate := range certificates {
		c := certificate

		g.Go(func() error {
			secret, err := s.search(c.TLS, clusterID, c.Cert)
			if err != nil {
				return microerror.Mask(err)
			}

			m.Lock()
			defer m.Unlock()
			err = fillTLSFromSecret(c.TLS, secret, clusterID, c.Cert)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		return ClusterOperator{}, microerror.Mask(err)
	}

	return clusterOperator, nil
}

func (s *Searcher) SearchDraining(clusterID string) (Draining, error) {
	var draining Draining

	certificates := []struct {
		TLS  *TLS
		Cert Cert
	}{
		{TLS: &draining.NodeOperator, Cert: NodeOperatorCert},
	}

	g := &errgroup.Group{}
	m := sync.Mutex{}

	for _, certificate := range certificates {
		c := certificate

		g.Go(func() error {
			secret, err := s.search(c.TLS, clusterID, c.Cert)
			if err != nil {
				return microerror.Mask(err)
			}

			m.Lock()
			defer m.Unlock()
			err = fillTLSFromSecret(c.TLS, secret, clusterID, c.Cert)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		return Draining{}, microerror.Mask(err)
	}

	return draining, nil
}

func (s *Searcher) SearchMonitoring(clusterID string) (Monitoring, error) {
	var monitoring Monitoring

	certificates := []struct {
		TLS  *TLS
		Cert Cert
	}{
		{TLS: &monitoring.Prometheus, Cert: PrometheusCert},
	}

	g := &errgroup.Group{}
	m := sync.Mutex{}

	for _, certificate := range certificates {
		c := certificate

		g.Go(func() error {
			secret, err := s.search(c.TLS, clusterID, c.Cert)
			if err != nil {
				return microerror.Mask(err)
			}

			m.Lock()
			defer m.Unlock()
			err = fillTLSFromSecret(c.TLS, secret, clusterID, c.Cert)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		return Monitoring{}, microerror.Mask(err)
	}

	return monitoring, nil
}

func (s *Searcher) SearchTLS(clusterID string, cert Cert) (TLS, error) {
	tls := &TLS{}

	secret, err := s.search(tls, clusterID, cert)
	if err != nil {
		return TLS{}, microerror.Mask(err)
	}

	err = fillTLSFromSecret(tls, secret, clusterID, cert)
	if err != nil {
		return TLS{}, microerror.Mask(err)
	}

	return *tls, nil
}

func (s *Searcher) search(tls *TLS, clusterID string, cert Cert) (*corev1.Secret, error) {
	// Select only secrets that match the given certificate and the given
	// cluster clusterID.
	selector := fmt.Sprintf("%s=%s, %s=%s", legacyCertificateLabel, cert, legacyClusterIDLabel, clusterID)

	watcher, err := s.k8sClient.CoreV1().Secrets(SecretNamespace).Watch(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	defer watcher.Stop()

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return nil, microerror.Maskf(executionFailedError, "watching secrets, selector = %q: unexpected closed channel", selector)
			}

			switch event.Type {
			case watch.Added:
				secret, ok := event.Object.(*corev1.Secret)
				if !ok || secret == nil {
					return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", secret, event.Object)
				}

				return secret, nil
			case watch.Deleted:
				// Noop. Ignore deleted events. These are
				// handled by the certificate operator.
			case watch.Error:
				return nil, microerror.Maskf(executionFailedError, "watching secrets, selector = %q: %v", selector, apierrors.FromObject(event.Object))
			}
		case <-time.After(s.watchTimeout):
			return nil, microerror.Maskf(timeoutError, "waiting secrets, selector = %q", selector)
		}
	}
}

func fillTLSFromSecret(tls *TLS, secret *corev1.Secret, clusterID string, cert Cert) error {
	gotClusterID := secret.Labels[legacyClusterIDLabel]
	if clusterID != gotClusterID {
		return microerror.Maskf(invalidSecretError, "expected clusterID = %q, got %q", clusterID, gotClusterID)
	}
	gotcert := secret.Labels[legacyCertificateLabel]
	if string(cert) != gotcert {
		return microerror.Maskf(invalidSecretError, "expected certificate = %q, got %q", cert, gotcert)
	}

	var ok bool
	if tls.CA, ok = secret.Data["ca"]; !ok {
		return microerror.Maskf(invalidSecretError, "%q key missing", "ca")
	}
	if tls.Crt, ok = secret.Data["crt"]; !ok {
		return microerror.Maskf(invalidSecretError, "%q key missing", "crt")
	}
	if tls.Key, ok = secret.Data["key"]; !ok {
		return microerror.Maskf(invalidSecretError, "%q key missing", "key")
	}

	return nil
}
