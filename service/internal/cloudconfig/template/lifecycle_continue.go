package template

const MasterInstanceLifecycleContinueService = `
[Unit]
Description=master-instance-lifecycle-continue job
After=k8s-kubelet.service 
Requires=k8s-kubelet.service 
[Service]
Type=simple
Restart=always
RestartSec=15
ExecStartPre=/opt/bin/etcd-healthcheck
ExecStart=/opt/bin/master-instance-lifecycle-continue
[Install]
WantedBy=multi-user.target
`

const MasterInstanceLifecycleContinue = `#!/bin/bash
set -o nounset
set -o pipefail

# AWS Metadata
export INSTANCEID=$(curl -s http://169.254.169.254/latest/meta-data/instance-id 2> /dev/null)

# AWS Autoscaling Group Name
export AUTOSCALINGGROUP=$(docker run --rm {{ .RegistryDomain }}/giantswarm/awscli:2.7.35 autoscaling describe-auto-scaling-instances --instance-ids=$INSTANCEID --query 'AutoScalingInstances[*].AutoScalingGroupName' --output text)

output=$(docker run --rm -i {{ .RegistryDomain }}/giantswarm/awscli:2.7.35 autoscaling complete-lifecycle-action --auto-scaling-group-name $AUTOSCALINGGROUP --lifecycle-hook-name ControlPlaneLaunching --instance-id $INSTANCEID --lifecycle-action-result CONTINUE 2>&1 > /dev/null)

# We ignore the following error: An error occurred (ValidationError) when calling the CompleteLifecycleAction operation: No active Lifecycle Action found with instance ID i-0f9f9f9f9f9f9f9f9
# This happens when the lifecycle hook is already completed.
if [[ $output != *"No active Lifecycle Action found"* ]]
then
    echo "Successfully completed lifecycle action. Master instance is ready receiving traffic."
    exit 0
else
    echo "Failed to complete lifecycle action. Got error: $output"
    exit 1
fi
`

const ETCDHealthCheck = `#!/bin/bash
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
