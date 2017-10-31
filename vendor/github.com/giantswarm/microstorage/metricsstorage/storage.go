package metricsstorage

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microstorage"
)

const (
	prometheusNamespace = "microstorage"

	putActionName    = "put"
	deleteActionName = "delete"
	existsActionName = "exists"
	listActionName   = "list"
	searchActionName = "search"
)

var (
	actionTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusNamespace,
			Name:      "action_total",
			Help:      "Total number of storage actions performed.",
		},
		[]string{"action"},
	)

	errorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusNamespace,
			Name:      "error_total",
			Help:      "Total number of storage actions that have errored.",
		},
		[]string{"action"},
	)

	actionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: prometheusNamespace,
			Name:      "action_duration_seconds",
			Help:      "Duration of time to perform storage actions.",
		},
		[]string{"action"},
	)
)

func init() {
	prometheus.MustRegister(actionTotal)
	prometheus.MustRegister(errorTotal)
	prometheus.MustRegister(actionDuration)
}

type Config struct {
	Underlying microstorage.Storage
}

func DefaultConfig() Config {
	return Config{
		Underlying: nil,
	}
}

type Storage struct {
	underlying microstorage.Storage
}

func New(config Config) (*Storage, error) {
	if config.Underlying == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Underlying must not be empty")
	}

	s := &Storage{
		underlying: config.Underlying,
	}

	return s, nil
}

func (s *Storage) Put(ctx context.Context, kv microstorage.KV) error {
	timer := prometheus.NewTimer(actionDuration.WithLabelValues(putActionName))
	defer timer.ObserveDuration()

	actionTotal.WithLabelValues(putActionName).Inc()

	err := s.underlying.Put(ctx, kv)

	if err != nil {
		errorTotal.WithLabelValues(putActionName).Inc()
	}

	return err
}

func (s *Storage) Delete(ctx context.Context, key microstorage.K) error {
	timer := prometheus.NewTimer(actionDuration.WithLabelValues(deleteActionName))
	defer timer.ObserveDuration()

	actionTotal.WithLabelValues(deleteActionName).Inc()

	err := s.underlying.Delete(ctx, key)

	if err != nil {
		errorTotal.WithLabelValues(deleteActionName).Inc()
	}

	return err
}

func (s *Storage) Exists(ctx context.Context, key microstorage.K) (bool, error) {
	timer := prometheus.NewTimer(actionDuration.WithLabelValues(existsActionName))
	defer timer.ObserveDuration()

	actionTotal.WithLabelValues(existsActionName).Inc()

	b, err := s.underlying.Exists(ctx, key)

	if err != nil {
		errorTotal.WithLabelValues(existsActionName).Inc()
	}

	return b, err
}

func (s *Storage) List(ctx context.Context, key microstorage.K) ([]microstorage.KV, error) {
	timer := prometheus.NewTimer(actionDuration.WithLabelValues(listActionName))
	defer timer.ObserveDuration()

	actionTotal.WithLabelValues(listActionName).Inc()

	kvs, err := s.underlying.List(ctx, key)

	if err != nil {
		errorTotal.WithLabelValues(listActionName).Inc()
	}

	return kvs, err
}

func (s *Storage) Search(ctx context.Context, key microstorage.K) (microstorage.KV, error) {
	timer := prometheus.NewTimer(actionDuration.WithLabelValues(searchActionName))
	defer timer.ObserveDuration()

	actionTotal.WithLabelValues(searchActionName).Inc()

	kv, err := s.underlying.Search(ctx, key)

	if err != nil {
		errorTotal.WithLabelValues(searchActionName).Inc()
	}

	return kv, err
}
