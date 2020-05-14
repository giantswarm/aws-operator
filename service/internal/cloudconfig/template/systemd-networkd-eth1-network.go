package template

const SystemdNetworkdEth1Network = `
# disable DHCP
[Match]
Name=eth1
[Network]
DHCP=no
`
