package filelogger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	folder = "logs"
)

type Config struct {
	Backoff   backoff.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

type FileLogger struct {
	backoff   backoff.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

func New(config Config) (*FileLogger, error) {
	if config.Backoff == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Backoff must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	f := &FileLogger{
		backoff:   config.Backoff,
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return f, nil
}

func (f FileLogger) StartLoggingPod(namespace, name string) error {
	req := f.k8sClient.CoreV1().RESTClient().Get().Namespace(namespace).Name(name).Resource("pods").SubResource("log").Param("follow", strconv.FormatBool(true))

	var readCloser io.ReadCloser
	var err error

	o := func() error {
		readCloser, err = req.Stream()
		if err != nil {
			return microerror.Mask(err)
		}
		return nil
	}
	n := backoff.NewNotifier(f.logger, context.Background())
	err = backoff.RetryNotify(o, f.backoff, n)
	if err != nil {
		return microerror.Mask(err)
	}

	go f.scan(readCloser, name)
	return nil
}

func (f FileLogger) scan(readCloser io.ReadCloser, name string) {
	defer readCloser.Close()

	logpath := filepath.Join(".", "logs")
	os.MkdirAll(logpath, os.ModePerm)

	logfilepath := path.Join(logpath, fmt.Sprintf("%s-logs.txt", name))

	outFile, err := os.Create(logfilepath)
	if err != nil {
		f.logger.Log("level", "error", "stack", microerror.Mask(err))
		return
	}

	defer outFile.Close()

	f.logger.Log("level", "debug", "message", fmt.Sprintf("logging output of %#q to %#q", name, outFile.Name()))
	_, err = io.Copy(outFile, readCloser)
	if err != nil {
		f.logger.Log("level", "error", "stack", microerror.Mask(err))
	}

	f.logger.Log("level", "debug", "message", fmt.Sprintf("logged output of %#q to %#q", name, outFile.Name()))
}
