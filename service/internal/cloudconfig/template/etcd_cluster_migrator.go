package template

const EtcdClusterMigratorService = `
[Unit]
Description=Install etcd-cluster-migrator
After=k8s-kubelet.service
Requires=k8s-kubelet.service

[Service]
Type=oneshot
ExecStart=/opt/bin/install-etcd-cluster-migrator

[Install]
WantedBy=multi-user.target
`

const EtcdClusterMigratorInstaller = `#!/bin/bash

export KUBECONFIG=/etc/kubernetes/kubeconfig/addons.yaml

for manifest in "etcd-cluster-migrator.yaml"
do
    while
        /opt/bin/kubectl apply -f /srv/$manifest
        [ "$?" -ne "0" ]
    do
        echo "failed to apply /srv/$manifest, retrying in 10 sec"
        sleep 10s
    done
done
`

const EtcdClusterMigratorManifest = `
apiVersion: v1
data:
  values: |
    app:
      baseDomain: {{.BaseDomain}}
    image:
      registry: {{.RegistryDomain}}
kind: ConfigMap
metadata:
  name: etcd-cluster-migrator-chart-values
  namespace: giantswarm
---
apiVersion: application.giantswarm.io/v1alpha1
kind: Chart
metadata:
  annotations:
    chart-operator.giantswarm.io/force-helm-upgrade: "true"
  labels:
    app: etcd-cluster-migrator
    chart-operator.giantswarm.io/version: 1.0.0
    giantswarm.io/organization: giantswarm
    giantswarm.io/service-type: managed
  name: etcd-cluster-migrator
  namespace: giantswarm
spec:
  config:
    configMap:
      name: etcd-cluster-migrator-chart-values
      namespace: giantswarm
      resourceVersion: ""
  name: etcd-cluster-migrator
  namespace: kube-system
  tarballURL: https://giantswarm.github.io/giantswarm-playground-test-catalog/etcd-cluster-migrator-0.0.0-e6ee2da93b136339bf927a610ed0c8b69a1c9428.tgz
  version: 0.0.0
`
