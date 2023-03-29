package template

// Startlimitburst is set to 70 because the hook has a timelimit of 17 minutes. (60/15sec retry)*17 = 68
// StartLimitIntervalSec is set to 17 minutes as anyways the hook will be sent automatically
const MasterInstanceLifecycleContinueService = `
[Unit]
Description=master-instance-lifecycle-continue job
Requires=network.target
After=network.target
StartLimitIntervalSec=1020
StartLimitBurst=68
[Service]
Type=simple
Restart=on-failure
RestartSec=15
ExecStartPre=/opt/bin/etcd-healthcheck
ExecStart=/opt/bin/master-instance-lifecycle-continue
[Install]
WantedBy=multi-user.target
`

const MasterInstanceLifecycleContinue = `#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

# AWS Metadata
export INSTANCEID=$(curl -s http://169.254.169.254/latest/meta-data/instance-id 2> /dev/null)

# AWS Autoscaling Group Name
export AUTOSCALINGGROUP=$(docker run --rm {{ .RegistryDomain }}/giantswarm/awscli:2.7.35 autoscaling describe-auto-scaling-instances --instance-ids=$INSTANCEID --query 'AutoScalingInstances[*].AutoScalingGroupName' --output text)

output=$(docker run --rm -i {{ .RegistryDomain }}/giantswarm/awscli:2.7.35 autoscaling complete-lifecycle-action --auto-scaling-group-name $AUTOSCALINGGROUP --lifecycle-hook-name ControlPlaneLaunching --instance-id $INSTANCEID --lifecycle-action-result CONTINUE 2>&1 > /dev/null)

if [ $? == 0 ]; then
    echo "Successfully completed lifecycle action. Master instance is ready receiving traffic."
    exit 0
else
    echo "Failed to complete lifecycle action. Got error: $output"
    exit 1
fi
`

const ETCDHealthCheck = `#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

statuscode=$(ETCDCTL_API=3 etcdctl \
    --cert /etc/kubernetes/ssl/etcd/client-crt.pem \
    --key /etc/kubernetes/ssl/etcd/client-key.pem \
    --cacert /etc/kubernetes/ssl/etcd/client-ca.pem \
    --endpoints "127.0.0.1:2379" endpoint health --write-out json | jq .[].health)

if [ $statuscode = "true" ]; then 
  echo "ETCD is healthy."
  exit 0
fi

echo "ETCD is not healthy yet."
exit 1
`
