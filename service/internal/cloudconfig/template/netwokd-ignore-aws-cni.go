package template

const NetworkdIgnoreAWSCNiInterfaces = `
[Match]
Name=%s

[Link]
Unamanaged=yes
`
