package helmclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
)

// PullChartTarball downloads a tarball from the provided tarball URL,
// returning the file path.
func (c *Client) PullChartTarball(ctx context.Context, tarballURL string) (string, error) {
	req, err := c.newRequest("GET", tarballURL)
	if err != nil {
		return "", microerror.Mask(err)
	}

	chartTarballPath, err := c.doFile(ctx, req)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return chartTarballPath, nil
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
			// Github pages 404 produces full HTML page which
			// obscures the logs.
			if resp.StatusCode == http.StatusNotFound {
				return microerror.Maskf(executionFailedError, fmt.Sprintf("got StatusCode %d for url %#q", resp.StatusCode, req.URL.String()))
			}
			return microerror.Maskf(executionFailedError, fmt.Sprintf("got StatusCode %d for url %#q with body %s", resp.StatusCode, req.URL.String(), buf.String()))
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

	b := backoff.NewMaxRetries(10, 5*time.Second)
	n := backoff.NewNotifier(c.logger, ctx)

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return tmpFileName, nil
}

func (c *Client) newRequest(method, url string) (*http.Request, error) {
	var buf io.Reader

	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}
