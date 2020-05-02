package template

const SystemdNetworkdEth1Network = `
# ensure that traffic arriving on eth1 leaves again from eth1 to prevent asymetric routing
[Match]
Name=eth1
[Network]
Address={{.MasterENIAddress}}/{{.MasterENISubnetSize}}

[RoutingPolicyRule]
Table=2
From={{.MasterENIAddress}}/32

[Route]
Destination=0.0.0.0/0
Gateway={{.MasterENIGateway}}
Table=2
`
