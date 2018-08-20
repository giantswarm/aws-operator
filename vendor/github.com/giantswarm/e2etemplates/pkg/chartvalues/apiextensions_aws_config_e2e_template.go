package chartvalues

const apiExtensionsAWSConfigE2ETemplate = `commonDomain: {{ .CommonDomain }}
clusterName: {{ .ClusterName }}
clusterVersion: v_0_1_0
sshPublicKey: {{ .SSHPublicKey }}
versionBundleVersion: {{ .VersionBundleVersion }}
aws:
  networkCIDR: "10.12.0.0/24"
  privateSubnetCIDR: "10.12.0.0/25"
  publicSubnetCIDR: "10.12.0.128/25"
  region: {{ .AWS.Region }}
  apiHostedZone: {{ .AWS.APIHostedZone }}
  ingressHostedZone: {{ .AWS.IngressHostedZone }}
  routeTable0: {{ .AWS.RouteTable0 }}
  routeTable1: {{ .AWS.RouteTable1 }}
  vpcPeerId: {{ .AWS.VPCPeerID }}
`
