package chartvalues

const apiExtensionsChartE2ETemplate = `
chart:
  name: "{{ .Chart.Name }}"
  namespace: "{{ .Chart.Namespace }}"
  config:
    configMap:
      name: "{{ .Chart.Config.ConfigMap.Name }}"
      namespace: "{{ .Chart.Config.ConfigMap.Namespace }}"
    secret:
      name: "{{ .Chart.Config.Secret.Name }}"
      namespace: "{{ .Chart.Config.Secret.Namespace }}"
  tarballURL: "{{ .Chart.TarballURL }}"

chartOperator:
  version: "{{ .ChartOperator.Version }}"

configMap:
  values: 
    {{ .ConfigMap.ValuesYAML }}

namespace: "{{ .Namespace }}"

secret:
  values: 
    {{ .Secret.ValuesYAML }}`
