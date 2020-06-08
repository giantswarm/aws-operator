package template

	// TODO we install etcd-cluster-migrator in every case of HA masters. The etcd-cluster-migrator app 
	// does not have negative effects on Tenant Clusters that were already created using the HA masters 
	// setup. Already migrated Tenant Clusters can also safely run this app for the time being. The 
	// workaround here for now is only so we don't have to spent too much time implementing a proper
	// managed app via our app catalogue, which only deploys the etcd-cluster-migrator on demand in 
	// case a Tenant Cluster is migrating automatically from 1 to 3 masters. See also the TODO issue below.
	// 
	//     https://github.com/giantswarm/giantswarm/issues/11397
	//
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
  tarballURL: https://giantswarm.github.io/giantswarm-playground-test-catalog/etcd-cluster-migrator-0.0.0-52e453cd5007181161e47ee079137debed053780.tgz
  version: 0.0.0
`
