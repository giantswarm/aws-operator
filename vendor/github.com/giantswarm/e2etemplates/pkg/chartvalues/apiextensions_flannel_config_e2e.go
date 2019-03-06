package chartvalues

var apiExtensionsFlannelConfigE2ETemplate = `
clusterName: "{{.ClusterID}}"
versionBundleVersion: "0.2.0"
flannel:
  network: "{{.Network}}"
  vni: {{.VNI}}
`
