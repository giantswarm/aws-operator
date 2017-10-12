#!/bin/sh

set -x

. ./env.sh

kubectl delete aws -l clusterID=${CLUSTER_NAME}

kubectl delete -f configmap.yaml
kubectl delete -f configmap-ssh.yaml
kubectl delete -f deployment.yaml
kubectl delete -f cluster.yaml

kubectl config unset clusters.${CLUSTER_NAME}
kubectl config unset contexts.${CLUSTER_NAME}
kubectl config unset users.${CLUSTER_NAME}