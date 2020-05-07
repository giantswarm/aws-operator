package template

const TemplateMainENI = `
{{- define "eni" -}}
{{ range .ENI.List }}
  {{ .Resource }}:
    Type: AWS::EC2::NetworkInterface
    Properties:
       Description: A Network interface used for etcd.
       GroupSet:
       - {{ .SecurityGroupID }}
       SubnetId: {{ .SubnetID }}
       PrivateIpAddress: {{ .IpAddress }}
       Tags:
       - Key: Name
         Value: {{ .Name }}
       - Key: node.k8s.amazonaws.com/no_manage
         Value: "true"
{{- end -}}
{{- end -}}
`
