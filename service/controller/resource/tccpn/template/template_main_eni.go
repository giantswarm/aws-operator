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
       PrivateIpAddress: {{ .ENI.IpAddress }}
       Tags:
       - Key: Name
         Value: {{ .ENI.Name }}
{{- end -}}
`
