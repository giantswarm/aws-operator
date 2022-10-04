package template

const MasterInstanceLifeCycleCompletionService = `
[Unit]
Description=master-instance-lifecyle-completion job
After=k8s-kubelet.service k8s-setup-network-env.service
Requires=k8s-kubelet.service k8s-setup-network-env.service
[Service]
ExecStartPre=/bin/bash -c "while ! /opt/bin/master-instance-healthcheck ; do sleep 1; done"
ExecStart=/opt/bin/master-instance-lifecycle-completion
[Install]
WantedBy=multi-user.target
`
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
Description=Execute master-instance-healthcheck every 5 minutes
[Timer]
OnCalendar=*:0/5
[Install]
WantedBy=multi-user.target
`

const MasterInstanceHealthCheck = `#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

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

const MasterInstanceLifeCycle = `#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

# AWS Metadata
export INSTANCEID=$(curl -s http://169.254.169.254/latest/meta-data/instance-id 2> /dev/null)
# AWS Autoscaling Group Name
export AUTOSCALINGGROUP=$(docker run --rm {{ .RegistryDomain }}/giantswarm/awscli:2.7.35 autoscaling describe-auto-scaling-instances --instance-ids=$INSTANCEID --query 'AutoScalingInstances[*].AutoScalingGroupName' --output text)

output=$(docker run --rm -i {{ .RegistryDomain }}/giantswarm/awscli:2.7.35 autoscaling complete-lifecycle-action --auto-scaling-group-name $AUTOSCALINGGROUP --lifecycle-hook-name ControlPlaneLaunching --instance-id $INSTANCEID --lifecycle-action-result CONTINUE)
if [ $? -eq 0 ]; then
    exit 0
fi

if [[ $output == *"No active Lifecycle Action found"* ]];
    echo "Successfully completed lifecycle action. Master instance is ready receiving traffic."
    exit 0
else
    echo "Failed to complete lifecycle action. Got error: $output"
    exit 1
fi
`
