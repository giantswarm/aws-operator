package filelogger

import (
	"context"
	"fmt"
	"io"
	"os"
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
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

type FileLogger struct {
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

func New(config Config) (*FileLogger, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	f := &FileLogger{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return f, nil
}

func (f *FileLogger) EnsurePodLogging(ctx context.Context, namespace, name string) error {
	var err error

	logDir := filepath.Join(".", "logs")
	logFile := fmt.Sprintf("%s-%s-logs.txt", namespace, name)
	logFilePath := filepath.Join(logDir, logFile)

	{
		f.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding if pod %#q in namespace %#q is logging", name, namespace))

		_, err = os.Stat(logFilePath)
		if os.IsNotExist(err) {
			f.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding pod %#q in namespace %#q is not logging", name, namespace))
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			f.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found pod %#q in namespace %#q is already logging", name, namespace))
			return nil
		}
	}

	{
		f.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring directory %#q exists", logDir))

		err = os.MkdirAll(logDir, os.ModePerm)
		if err != nil {
			return microerror.Mask(err)
		}

		f.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured directory %#q exists", logDir))
	}

	var logStream io.ReadCloser
	{
		f.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating log stream for pod %#q in namespace %#q", name, namespace))

		req := f.k8sClient.CoreV1().RESTClient().Get().Namespace(namespace).Name(name).Resource("pods").SubResource("log").Param("follow", strconv.FormatBool(true))

		o := func() error {
			logStream, err = req.Stream()
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}
		b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)
		n := backoff.NewNotifier(f.logger, ctx)
		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}

		f.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created log stream for pod %#q in namespace %#q", name, namespace))
	}

	{
		f.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("starting logging job for pod %#q in namespace %#q", name, namespace))

		go f.scan(ctx, logStream, logFilePath)

		f.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("started logging job for pod %#q in namespace %#q", name, namespace))
	}
	return nil
}

func (f *FileLogger) scan(ctx context.Context, readCloser io.ReadCloser, logFilePath string) {
	defer readCloser.Close()

	outFile, err := os.Create(logFilePath)
	if err != nil {
		f.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("failed to create file %#q", logFilePath), "stack", fmt.Sprintf("%#v", err))
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, readCloser)
	if err != nil {
		f.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("failed write log stream to file %#q", logFilePath), "stack", fmt.Sprintf("%#v", err))
	}
}
