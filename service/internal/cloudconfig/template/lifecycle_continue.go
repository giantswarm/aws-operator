package template

const MasterInstanceLifecycleContinueService = `
[Unit]
Description=master-instance-lifecycle-continue job
After=k8s-kubelet.service etcd3.service
Requires=k8s-kubelet.service etcd3.service
[Service]
Type=simple
Restart=on-failure
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

output=$(docker run --rm -i {{ .RegistryDomain }}/giantswarm/awscli:2.7.35 autoscaling complete-lifecycle-action --auto-scaling-group-name $AUTOSCALINGGROUP --lifecycle-hook-name ControlPlaneLaunching --instance-id $INSTANCEID --lifecycle-action-result CONTINUE)
if [ $? -eq 0 ]; then
    exit 0
fi

# We ignore the following error: An error occurred (ValidationError) when calling the CompleteLifecycleAction operation: No active Lifecycle Action found with instance ID i-0f9f9f9f9f9f9f9f9
# This happens when the lifecycle hook is already completed.
if [[ $output == *"No active Lifecycle Action found"* ]];
    echo "Successfully completed lifecycle action. Master instance is ready receiving traffic."
    exit 0
else
    echo "Failed to complete lifecycle action. Got error: $output"
    exit 1
fi
`
