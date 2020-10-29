ln -s /go/pkg/mod/$(cat go.mod | grep k8scloudconfig | awk '{print $1"@"$2}') /opt/ignition

dlv debug --headless --listen=:2345 --log --api-version=2 -- daemon --config.dirs=/var/run/aws-operator/configmap/ --config.dirs=/var/run/aws-operator/secret/ --config.files=config --config.files=secret
