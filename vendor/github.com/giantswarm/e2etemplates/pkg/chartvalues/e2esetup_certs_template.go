package chartvalues

const e2eSetupCertsTemplate = `
cluster:
  id: {{ .Cluster.ID }}
commonDomain: {{ .CommonDomain }}
ipSans:
  - "172.31.0.1"
organizations:
  - "system:masters"
`
