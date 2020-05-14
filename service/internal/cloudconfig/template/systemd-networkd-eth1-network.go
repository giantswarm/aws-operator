package template

const SystemdNetworkdEth1Network = `
# disable DHCP
[Match]
Name=eth1
[Network]
DHCP=no
Address=127.0.0.1/32
`
