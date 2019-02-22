package chartvalues

const apiExtensionsAppE2ETemplate = `
app:
  name: "{{ .App.Name }}"
  namespace: "{{ .App.Namespace }}"
  catalog: "{{ .App.Catalog }}"
  config:
    configMap:
      name: "{{ .App.Config.ConfigMap.Name }}"
      namespace: "{{ .App.Config.ConfigMap.Namespace }}"
    secret:
      name: "{{ .App.Config.Secret.Name }}"
      namespace: "{{ .App.Config.Secret.Namespace }}"
  version: "{{ .App.Version }}"

appCatalog:
  name: "{{ .AppCatalog.Name }}"
  title: "{{ .AppCatalog.Title }}"
  description: "{{ .AppCatalog.Description }}"
  logoURL: "{{ .AppCatalog.LogoURL }}"
  storage: 
    type: "{{ .AppCatalog.Storage.Type }}" 
    url: "{{ .AppCatalog.Storage.URL }}" 

appOperator:
  version: "{{ .AppOperator.Version }}"

configMap:
  values: 
    {{ .ConfigMap.ValuesYAML }}

namespace: "{{ .Namespace }}"

secret:
  values: 
    {{ .Secret.ValuesYAML }}`
