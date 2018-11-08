package chartvalues

const apiExtensionsAWSConfigE2ETemplate = `commonDomain: {{ .CommonDomain }}
clusterName: {{ .ClusterName }}
clusterVersion: v_0_1_0
sshPublicKey: {{ .SSHPublicKey }}
versionBundleVersion: {{ .VersionBundleVersion }}
aws:
  networkCIDR: {{ .AWS.NetworkCIDR }}
  privateSubnetCIDR: {{ .AWS.PrivateSubnetCIDR }}
  publicSubnetCIDR: {{ .AWS.PublicSubnetCIDR }}
  region: {{ .AWS.Region }}
  apiHostedZone: {{ .AWS.APIHostedZone }}
  ingressHostedZone: {{ .AWS.IngressHostedZone }}
  routeTable0: {{ .AWS.RouteTable0 }}
  routeTable1: {{ .AWS.RouteTable1 }}
  vpcPeerId: {{ .AWS.VPCPeerID }}
`
