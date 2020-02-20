package cloudconfig


const SystemdNetworkdEth1Network = `
# ensure that traffic arriving on eth1 leaves again from eth1 to prevent asymetric routing
[Match]
Name=eth1
[Network]
Address={{index .MasterENIAddress .MasterID}}/{{.MasterENISubnetSize}}

[RoutingPolicyRule]
Table=2
From={{index .MasterENIAddress .MasterID}}/32

[Route]
Destination=0.0.0.0/0
Gateway={{index .MasterENIGateway .MasterID}}
Table=2
`