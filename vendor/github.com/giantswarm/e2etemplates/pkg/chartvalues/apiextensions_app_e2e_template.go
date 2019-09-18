package chartvalues

const apiExtensionsAppE2ETemplate = `
apps:
  - name: "{{ .App.Name }}"
    namespace: "{{ .App.Namespace }}"
    catalog: "{{ .App.Catalog }}"
{{- if or .App.Config.ConfigMap.Name .App.Config.Secret.Name }}
    config:
{{- if .App.Config.ConfigMap.Name }}
      configMap:
        name: "{{ .App.Config.ConfigMap.Name }}"
        namespace: "{{ .App.Config.ConfigMap.Namespace }}"
{{- end }}
{{- if .App.Config.Secret.Name }}
      secret:
        name: "{{ .App.Config.Secret.Name }}"
        namespace: "{{ .App.Config.Secret.Namespace }}"
{{- end }}
{{- end }}
{{- if .App.KubeConfig }}
    kubeConfig:
      inCluster: {{ .App.KubeConfig.InCluster }}
{{- if not .App.KubeConfig.InCluster }}
      secret:
        name: "{{ .App.KubeConfig.Secret.Name }}"
        namespace: "{{ .App.KubeConfig.Secret.Namespace }}"
{{- end }}
{{- end }}
    version: "{{ .App.Version }}"
  # Added app CR for bootstrapping chart-operator
  - name: "chart-operator"
    namespace: "giantswarm"
    catalog: "giantswarm-catalog"
    kubeConfig:
      inCluster: true
    version: "0.9.0"

appCatalogs:
  - name: "{{ .AppCatalog.Name }}"
    title: "{{ .AppCatalog.Title }}"
    description: "{{ .AppCatalog.Description }}"
    logoURL: "{{ .AppCatalog.LogoURL }}"
    storage:
      type: "{{ .AppCatalog.Storage.Type }}"
      url: "{{ .AppCatalog.Storage.URL }}"
  - name: "giantswarm-catalog"
    title: "giantswarm-catalog"
    description: "giantswarm catalog"
    logoUrl: "http://giantswarm.com/catalog-logo.png"
    storage:
      type: "helm"
      url: "https://giantswarm.github.com/giantswarm-catalog/"

appOperator:
  version: "{{ .AppOperator.Version }}"

{{ if .App.Config.ConfigMap.Name -}}
configMaps:
  {{ .App.Config.ConfigMap.Name }}:
    {{ .ConfigMap.ValuesYAML }}
{{- end }}

namespace: "{{ .Namespace }}"

{{ if .App.Config.Secret.Name -}}
secrets:
  {{ .App.Config.Secret.Name }}:
    {{ .Secret.ValuesYAML }}
{{- end }}`
