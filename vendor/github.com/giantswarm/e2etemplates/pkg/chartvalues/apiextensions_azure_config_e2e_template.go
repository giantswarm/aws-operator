package chartvalues

var apiExtensionsAzureConfigE2ETemplate = `
azure:
  calicoSubnetCIDR: {{ .Azure.CalicoSubnetCIDR }}
  cidr: {{ .Azure.CIDR }}
  location: {{ .Azure.Location }}
  masterSubnetCIDR: {{ .Azure.MasterSubnetCIDR }}
  vmSizeMaster: {{ .Azure.VMSizeMaster }}
  vmSizeWorker: {{ .Azure.VMSizeWorker }}
  vpnSubnetCIDR: {{ .Azure.VPNSubnetCIDR }}
  workerSubnetCIDR: {{ .Azure.WorkerSubnetCIDR }}
clusterName: {{ .ClusterName }}
commonDomain: {{ .CommonDomain }}
commonDomainResourceGroup: {{ .CommonDomainResourceGroup }}
versionBundleVersion: {{ .VersionBundleVersion }}
`
