package localkube

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/giantswarm/microerror"
)

const (
	minikubeDownloadURL = "https://github.com/kubernetes/minikube/releases/download/$MINIKUBE_VERSION/minikube-linux-amd64"
)

type Localkube struct {
}

func New() *Localkube {
	return &Localkube{}
}

func (l *Localkube) SetUp() error {
	// download minikube binary.
	if err := downloadFromURL(minikubeDownloadURL); err != nil {
		return microerror.Mask(err)
	}

	commands := []string{
		"chmod a+x ./minikube-linux-amd64",
		"sudo ./minikube-linux-amd64 start --vm-driver=none --extra-config=apiserver.Authorization.Mode=RBAC",
		"sudo chown -R $USER $HOME/.kube",
		"sudo chgrp -R $USER $HOME/.kube",
		"sudo chown -R $USER $HOME/.minikube",
		"sudo chgrp -R $USER $HOME/.minikube",
		"./minikube-linux-amd64 update-context",
	}
	for _, command := range commands {
		if err := l.runCmd(command); err != nil {
			return microerror.Mask(err)
		}
	}
	return nil
}

func (l *Localkube) runCmd(command string) error {
	command = os.ExpandEnv(command)
	items := strings.Fields(command)
	cmd := exec.Command(items[0], items[1:]...)
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func downloadFromURL(url string) error {
	url = os.ExpandEnv(url)

	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	output, err := os.Create(fileName)
	if err != nil {
		return microerror.Mask(err)
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		return microerror.Mask(err)
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
