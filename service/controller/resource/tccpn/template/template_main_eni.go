package template

const TemplateMainENI = `
{{- define "eni" -}}
{{- range $i,$e := .ENI.ENIs }}
  {{ $e.ResourceName }}:
    Type: AWS::EC2::NetworkInterface
    Properties:
       Description: A Network interface used for etcd{{ $i }}.
       GroupSet:
       - {{ $e.SecurityGroupID }}
       SubnetId: {{ $e.SubnetID }}
       PrivateIpAddress: {{ $e.IpAddress }}
       Tags:
       - Key: Name
         Value: {{ $e.Name }}
       - Key: node.k8s.amazonaws.com/no_manage
         Value: "true"
{{- end -}}
{{- end -}}
`
