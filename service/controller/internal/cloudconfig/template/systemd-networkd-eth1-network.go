package cloudconfig


const SystemdNetworkdEth1Network = `
# ensure that traffic arriving on eth1 leaves again from eth1 to prevent asymetric routing
[Match]
Name=eth1
[Network]
Address={{index .MasterENIAddresses .MasterID}}/{{.MasterENISubnetSize}}

[RoutingPolicyRule]
Table=2
From={{index .MasterENIAddresses .MasterID}}/32

[Route]
Destination=0.0.0.0/0
Gateway={{index .MasterENIGateways .MasterID}}
Table=2
`