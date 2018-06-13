package e2etemplates

const ApiextensionsAWSConfigE2EChartValues = `commonDomain: ${COMMON_DOMAIN}
clusterName: ${CLUSTER_NAME}
clusterVersion: v_0_1_0
sshPublicKey: ${IDRSA_PUB}
versionBundleVersion: ${VERSION_BUNDLE_VERSION}
aws:
  networkCIDR: "10.12.0.0/24"
  privateSubnetCIDR: "10.12.0.0/25"
  publicSubnetCIDR: "10.12.0.128/25"
  region: ${AWS_REGION}
  apiHostedZone: ${AWS_API_HOSTED_ZONE_GUEST}
  ingressHostedZone: ${AWS_INGRESS_HOSTED_ZONE_GUEST}
  routeTable0: ${AWS_ROUTE_TABLE_0}
  routeTable1: ${AWS_ROUTE_TABLE_1}
  vpcPeerId: ${AWS_VPC_PEER_ID}
`
