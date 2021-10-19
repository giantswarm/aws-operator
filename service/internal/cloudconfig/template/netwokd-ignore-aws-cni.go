package template

const NetworkdIgnoreAWSCNiInterfaces = `
[Match]
Name=%s

[Link]
Unmanaged=yes

[Network]
KeepConfiguration=yes
`
