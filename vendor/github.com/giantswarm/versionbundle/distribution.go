package versionbundle

import (
	"fmt"

	"github.com/coreos/go-semver/semver"
	"github.com/giantswarm/microerror"
)

type DistributionConfig struct {
	Bundles []Bundle
}

func DefaultDistributionConfig() DistributionConfig {
	return DistributionConfig{
		Bundles: nil,
	}
}

type Distribution struct {
	bundles []Bundle
	version string
}

func NewDistribution(config DistributionConfig) (Distribution, error) {
	if len(config.Bundles) == 0 {
		return Distribution{}, microerror.Maskf(invalidConfigError, "config.Bundles must not be empty")
	}

	version, err := aggregateDistributionVersion(config.Bundles)
	if err != nil {
		return Distribution{}, microerror.Maskf(invalidConfigError, err.Error())
	}

	d := Distribution{
		bundles: config.Bundles,
		version: version,
	}

	return d, nil
}

func (d Distribution) Bundles() []Bundle {
	return CopyBundles(d.bundles)
}

func (d Distribution) Version() string {
	return d.version
}

func aggregateDistributionVersion(bundles []Bundle) (string, error) {
	var major int64
	var minor int64
	var patch int64

	for _, b := range bundles {
		v, err := semver.NewVersion(b.Version)
		if err != nil {
			return "", microerror.Mask(err)
		}

		major += v.Major
		minor += v.Minor
		patch += v.Patch
	}

	version := fmt.Sprintf("%d.%d.%d", major, minor, patch)

	return version, nil
}
