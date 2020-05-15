package template

const TemplateMainENI = `
{{- define "eni" -}}
  MasterEni:
    Type: AWS::EC2::NetworkInterface
    Properties:
       Description: A Network interface used for etcd.
       GroupSet:
       - {{ .ENI.SecurityGroupID }}
       SubnetId: {{ .ENI.SubnetID }}
       Tags:
       - Key: Name
         Value: {{ .ENI.Name }}
       - Key: node.k8s.amazonaws.com/no_manage
         Value: "true"
{{- end -}}
`
