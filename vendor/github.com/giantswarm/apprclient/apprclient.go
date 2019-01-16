// Package apprclient holds the client code required to interact with a CNR
// backend.
package apprclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
)

type Payload struct {
	Release   string `json:"release"`
	MediaType string `json:"media_type"`
	Blob      string `json:"blob"`
}

type Response struct {
	Status string `json:"status"`
}

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
		config.Fs = afero.NewOsFs()
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

// DeleteRelease removes a release from the server.
func (c *Client) DeleteRelease(ctx context.Context, name, release string) error {
	p := path.Join("packages", c.organization, name, release, "helm")

	req, err := c.newRequest("DELETE", p)
	if err != nil {
		return microerror.Mask(err)
	}

	var r Response
	err = c.do(ctx, req, &r)
	if err != nil {
		return microerror.Mask(err)
	}

	if r.Status != deletedStatus {
		return microerror.Mask(unknownStatusError)
	}

	return nil
}

// GetReleaseVersion queries CNR for the release version of the chart
// represented by the given name and channel.
func (c *Client) GetReleaseVersion(ctx context.Context, name, channel string) (string, error) {
	p := path.Join("packages", c.organization, name, "channels", channel)

	req, err := c.newRequest("GET", p)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var ch cnrChannel
	err = c.do(ctx, req, &ch)

	if err != nil {
		return "", microerror.Mask(err)
	}

	return ch.Current, nil
}

// PromoteChart puts a release of the given chart in a channel.
func (c *Client) PromoteChart(ctx context.Context, name, release, channel string) error {
	p := path.Join("packages", c.organization, name, "channels", channel, release)

	req, err := c.newRequest("POST", p)
	if err != nil {
		return microerror.Mask(err)
	}

	ch := &cnrChannel{}
	err = c.do(ctx, req, ch)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// PullChartTarball downloads a tarball with the chart described by
// the given chart name and channel, returning the file path.
func (c *Client) PullChartTarball(ctx context.Context, name, channel string) (string, error) {
	release, err := c.GetReleaseVersion(ctx, name, channel)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return c.PullChartTarballFromRelease(ctx, name, release)
}

// PullChartTarballFromRelease downloads a tarball with the chart described
// by the given chart name and release, returning the file path.
func (c *Client) PullChartTarballFromRelease(ctx context.Context, name, release string) (string, error) {
	p := path.Join("packages", c.organization, name, release, "helm", "pull")

	req, err := c.newRequest("GET", p)
	if err != nil {
		return "", microerror.Mask(err)
	}

	chartTarballPath, err := c.doFile(ctx, req)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return chartTarballPath, nil
}

// PushChartTarball sends a tarball to the server to be installed for the given
// name and release.
func (c *Client) PushChartTarball(ctx context.Context, name, release, tarballPath string) error {
	p := path.Join("packages", c.organization, name)

	blob, err := c.readBlob(tarballPath)
	if err != nil {
		return microerror.Mask(err)
	}

	payload := &Payload{
		Release:   release,
		MediaType: "helm",
		Blob:      blob,
	}
	req, err := c.newPayloadRequest(p, payload)
	if err != nil {
		return microerror.Mask(err)
	}

	var r Response
	err = c.do(ctx, req, &r)
	if err != nil {
		return microerror.Mask(err)
	}

	if r.Status != okStauts {
		return microerror.Mask(unknownStatusError)
	}

	return nil
}

func (c *Client) newRequest(method, path string) (*http.Request, error) {
	u := &url.URL{Path: path}
	dest := c.base.ResolveReference(u)

	var buf io.Reader

	req, err := http.NewRequest(method, dest.String(), buf)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Client) newPayloadRequest(path string, payload *Payload) (*http.Request, error) {
	u := &url.URL{Path: path}
	dest := c.base.ResolveReference(u)

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	buf := bytes.NewReader(b)

	req, err := http.NewRequest("POST", dest.String(), buf)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) error {
	req = req.WithContext(ctx)

	o := func() error {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return microerror.Mask(err)
		}
		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(v)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)
	n := backoff.NewNotifier(c.logger, ctx)

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (c *Client) doFile(ctx context.Context, req *http.Request) (string, error) {
	var tmpFileName string

	req = req.WithContext(ctx)

	o := func() error {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return microerror.Mask(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(resp.Body)
			if err != nil {
				return microerror.Mask(err)
			}
			return microerror.Maskf(invalidStatusCodeError, fmt.Sprintf("got StatusCode %d with body %s", resp.StatusCode, buf.String()))
		}

		tmpfile, err := afero.TempFile(c.fs, "", "chart-tarball")
		if err != nil {
			return microerror.Mask(err)
		}
		defer tmpfile.Close()

		_, err = io.Copy(tmpfile, resp.Body)
		if err != nil {
			return microerror.Mask(err)
		}

		tmpFileName = tmpfile.Name()

		return nil
	}
	b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)
	n := backoff.NewNotifier(c.logger, ctx)

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return tmpFileName, nil
}

func (c *Client) readBlob(path string) (string, error) {
	afs := &afero.Afero{Fs: c.fs}

	content, err := afs.ReadFile(path)
	if err != nil {
		return "", microerror.Mask(err)
	}

	data := base64.StdEncoding.EncodeToString(content)

	return data, nil
}
