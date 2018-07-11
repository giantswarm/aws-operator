package guest

const RecordSets = `{{define "record_sets"}}
{{ if .Route53Enabled }}
  HostedZone:
    Type: 'AWS::Route53::HostedZone'
    Properties:
      Name: '{{ .ClusterID }}.k8s.{{ .BaseDomain }}.'
  ApiRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      AliasTarget:
        DNSName: !GetAtt ApiLoadBalancer.DNSName
        HostedZoneId: !GetAtt ApiLoadBalancer.CanonicalHostedZoneNameID
        EvaluateTargetHealth: false
      Name: 'api.{{ .ClusterID }}.k8s.{{ .BaseDomain }}.'
      HostedZoneId: !Ref 'HostedZone'
      Type: A
  EtcdRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      Name: 'etcd.{{ .ClusterID }}.k8s.{{ .BaseDomain }}.'
      HostedZoneId: !Ref 'HostedZone'
      TTL: '900'
      Type: CNAME
      ResourceRecords:
        - !GetAtt {{ .MasterInstanceResourceName }}.PrivateDnsName
  IngressRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      AliasTarget:
        DNSName: !GetAtt IngressLoadBalancer.DNSName
        HostedZoneId: !GetAtt IngressLoadBalancer.CanonicalHostedZoneNameID
        EvaluateTargetHealth: false
      Name: 'ingress.{{ .ClusterID }}.k8s.{{ .BaseDomain }}.'
      HostedZoneId: !Ref 'HostedZone'
      Type: A
  IngressWildcardRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      Name: '*.{{ .ClusterID }}.k8s.{{ .BaseDomain }}.'
      HostedZoneId: !Ref 'HostedZone'
      TTL: '900'
      Type: CNAME
      ResourceRecords:
        - !Ref 'IngressRecordSet'
{{end}}
{{end}}`
