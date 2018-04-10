// Package apprclient holds the client code required to interact with a CNR
// backend.
package apprclient

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
)

// Config represents the configuration used to create a appr client.
type Config struct {
	Fs     afero.Fs
	Logger micrologger.Logger

	Address      string
	Organization string
}

// Client knows how to talk with a CNR server.
type Client struct {
	fs         afero.Fs
	httpClient *http.Client
	logger     micrologger.Logger

	base         *url.URL
	organization string
}

// New creates a new configured appr client.
func New(config Config) (*Client, error) {
	if config.Fs == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Fs must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.Address == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Address must not be empty", config)
	}
	if config.Organization == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Organization must not be empty", config)
	}

	base, err := url.Parse(config.Address + "/cnr/api/v1/")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// set client timeout to prevent leakages.
	httpClient := &http.Client{
		Timeout: time.Second * httpClientTimeout,
	}

	c := &Client{
		fs:     config.Fs,
		logger: config.Logger,

		base:         base,
		httpClient:   httpClient,
		organization: config.Organization,
	}

	return c, nil
}

// GetReleaseVersion queries CNR for the release version of the chart
// represented by the given name and channel.
func (c *Client) GetReleaseVersion(name, channel string) (string, error) {
	p := path.Join("packages", c.organization, name, "channels", channel)

	req, err := c.newRequest("GET", p)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var ch Channel
	_, err = c.do(req, &ch)

	if err != nil {
		return "", microerror.Mask(err)
	}

	return ch.Current, nil
}

// PullChartTarball downloads a tarball with the chart described by the given
// chart name and channel, returning the file path.
func (c *Client) PullChartTarball(name, channel string) (string, error) {
	release, err := c.GetReleaseVersion(name, channel)
	if err != nil {
		return "", microerror.Mask(err)
	}

	p := path.Join("packages", c.organization, name, release, "helm", "pull")

	req, err := c.newRequest("GET", p)
	if err != nil {
		return "", microerror.Mask(err)
	}

	chartTarballPath, err := c.doFile(req)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return chartTarballPath, nil
}

func (c *Client) newRequest(method, path string) (*http.Request, error) {
	u := &url.URL{Path: path}
	dest := c.base.ResolveReference(u)

	var buf io.ReadWriter

	req, err := http.NewRequest(method, dest.String(), buf)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return resp, nil
}

func (c *Client) doFile(req *http.Request) (string, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", microerror.Mask(err)
	}
	defer resp.Body.Close()

	tmpfile, err := afero.TempFile(c.fs, "", "chart-tarball")
	if err != nil {
		return "", microerror.Mask(err)
	}
	defer tmpfile.Close()

	_, err = io.Copy(tmpfile, resp.Body)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return tmpfile.Name(), nil
}
