package e2etemplates

const ApiextensionsChartConfigE2EChartValues = `chart:
  channel: {{ .Channel }}
  configMap:
    name: {{ .ConfigMap.Name }}
    namespace: {{ .ConfigMap.Namespace }}
    resourceVersion: {{ .ConfigMap.ResourceVersion }}
  name: {{ .Name }}
  namespace: {{ .Namespace }}
  release: {{ .Release }}
  secret:
    name: {{ .Secret.Name }}
    namespace: {{ .Secret.Namespace }}
    resourceVersion: {{ .Secret.ResourceVersion }}
versionBundleVersion: {{ .VersionBundleVersion }}
`
