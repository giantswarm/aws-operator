package chartvalues

var apiExtensionsAzureConfigE2ETemplate = `
azure:
  {{ $length := len .Azure.AvailabilityZones }} {{- if gt $length 0 -}}
  availabilityZones:
  {{ range $index, $element := .Azure.AvailabilityZones -}}
  - {{ . }}
  {{ end -}}
  {{ end -}}
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
sshPublicKey: {{ .SSHPublicKey }}
sshUser: {{ .SSHUser }}
versionBundleVersion: {{ .VersionBundleVersion }}
`
