package versionbundle

import (
	"net/url"

	"github.com/giantswarm/microerror"
)

type Authority struct {
	Endpoint *URL   `yaml:"endpoint"`
	Name     string `yaml:"name"`
	Provider string `yaml:"provider"`
	Version  string `yaml:"version"`
}

// URL is a hack referring to the native url.URL in order to support yaml
// unmarshaling.
type URL struct {
	*url.URL
}

func (u *URL) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return microerror.Mask(err)
	}
	url, err := url.Parse(s)
	if err != nil {
		return microerror.Mask(err)
	}

	u.URL = url

	return nil
}
