package chartvalues

const apiExtensionsReleaseE2ETemplate = `active: {{ .Active }}
authorities:
  {{- range .Authorities }}
  - name: "{{ .Name }}"
    version: "{{ .Version }}"
  {{- end }}
date: {{ .Date }}
name: {{ .Name }}
namespace: {{ .Namespace }}
provider: {{ .Provider }}
version: {{ .Version }}
versionBundle:
  version: {{ .VersionBundle.Version }}
`
