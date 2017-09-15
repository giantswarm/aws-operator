package migrator

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
)

type Config struct {
	Logger micrologger.Logger
}

// DefaultConfig creates a new configuration with the default settings.
func DefaultConfig() Config {
	return Config{
		Logger: nil, // Required.
	}
}

type Migrator struct {
	logger micrologger.Logger
}

func New(config Config) (*Migrator, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger is empty")
	}

	m := &Migrator{
		logger: config.Logger,
	}

	return m, nil
}

func (m *Migrator) Migrate(ctx context.Context, dst, src microstorage.Storage) error {
	var err error

	m.logger.Log("debug", "listing all KVs")
	kvs, err := src.List(ctx, microstorage.RootKey)
	if microstorage.IsNotFound(err) {
		m.logger.Log("debug", "src sotrage is empty")
		return nil
	} else if err != nil {
		return microerror.Maskf(err, "src storage: listing key=/")
	}

	m.logger.Log("debug", fmt.Sprintf("transfering %d entries", len(kvs)))
	var migrated int
	for _, kv := range kvs {
		exists, err := dst.Exists(ctx, kv.K())
		if err != nil {
			return microerror.Maskf(err, "dst storage: checking key=%s", kv.Key())
		}

		if exists {
			continue
		}

		err = dst.Put(ctx, kv)
		if err != nil {
			return microerror.Maskf(err, "dst storage: putting key=%s", kv.Key())
		}
		migrated++
	}

	m.logger.Log("info", fmt.Sprintf("migrated %d/%d remaining entries", migrated, len(kvs)))
	return nil
}
