package template

const MasterInstanceHealtCheckService = `
[Unit]
Description=master-instance-healthcheck job
After=k8s-kubelet.service k8s-setup-network-env.service
Requires=k8s-kubelet.service k8s-setup-network-env.service
[Service]
Type=oneshot
ExecStart=/opt/bin/master-instance-healthcheck
[Install]
WantedBy=multi-user.target
`

const MasterInstanceHealtCheckTimer = `
[Unit]
Description=Execute master-instance-healthcheck every minute
[Timer]
OnCalendar=*-*-* *:*:00
[Install]
WantedBy=multi-user.target
`

const MasterInstanceHealthCheck = `#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

# Wait for 15 minutes uptime until we check
if test $(cut -d '.' -f1 /proc/uptime) -lt 900; then
  echo "Uptime is less than 15 minutes ago, skipping..."
  exit 0
fi

# AWS Metadata
export INSTANCEID=$(curl -s http://169.254.169.254/latest/meta-data/instance-id 2> /dev/null)

# Additional ENVs
. /etc/network-environment

# Color
export Red=$'\033[1;31m'
export Green=$'\033[1;32m'
export Yellow=$'\033[1;33m'
export NoColor=$'\033[0m'


function retry() {
  local max_attempts=${ATTEMPTS-3}
  local timeout=${TIMEOUT-15}
  local attempt=0
  local exitCode=0

  while [[ $attempt < $max_attempts ]]
  do
    "$@"
    exitCode=$?

    if [[ $exitCode == 0 ]]
    then
      break
    fi

    echo "${Red}Failed to get status from $2.${NoColor} Retrying in $timeout.. seconds" 1>&2
    sleep $timeout
    attempt=$(( attempt + 1 ))
    timeout=$(( timeout * 2 ))
  done

  if [[ $exitCode != 0 ]]
  then
    echo "Mark EC2 instance ${Yellow}$INSTANCEID${NoColor} as ${Red}UNHEALTHY${NoColor}"
    docker run --rm -i {{ .RegistryDomain }}/giantswarm/awscli:2.7.35 autoscaling set-instance-health --instance-id $INSTANCEID --health-status Unhealthy
    exit $exitCode
  fi

  echo "STATUS is ${Green}OK${NoColor} for $2"
  return $exitCode
}

# Checking etcd
function etcd_status {
  local statuscode

  statuscode=$(ETCDCTL_API=3 etcdctl \
    --cert /etc/kubernetes/ssl/etcd/client-crt.pem \
    --key /etc/kubernetes/ssl/etcd/client-key.pem \
    --cacert /etc/kubernetes/ssl/etcd/client-ca.pem \
    --endpoints "127.0.0.1:2379" endpoint health --write-out json | jq .[].health)

  if [ $statuscode = "true" ]; then
     return 0
  fi
  return 1
}

function kubelet_status {
  local statuscode

  statuscode=$(curl -k -LI https://$DEFAULT_IPV4:10250/metrics -o /dev/null -w '%{http_code}\n' -s)

  if [ $statuscode -eq 200 ]; then
     return 0
  fi
  return 1
}

function apiserver_status {
  local statuscode

  statuscode=$(curl -k https://127.0.0.1/healthz -o /dev/null -w '%{http_code}\n' -s)

  if [ $statuscode -eq 200 ]; then
    return 0
  fi
  return 1
}

retry etcd_status "ETCD" || true
retry kubelet_status "KUBELET" || true
retry apiserver_status "APISERVER" || true
`
