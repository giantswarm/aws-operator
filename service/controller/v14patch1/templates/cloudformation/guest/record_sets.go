package guest

const RecordSets = `{{define "record_sets"}}
{{- $v := .Guest.RecordSets }}
{{ if $v.Route53Enabled }}
  HostedZone:
    Type: 'AWS::Route53::HostedZone'
    Properties:
      Name: '{{ $v.ClusterID }}.k8s.{{ $v.BaseDomain }}.'
  ApiRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      AliasTarget:
        DNSName: !GetAtt ApiLoadBalancer.DNSName
        HostedZoneId: !GetAtt ApiLoadBalancer.CanonicalHostedZoneNameID
        EvaluateTargetHealth: false
      Name: 'api.{{ $v.ClusterID }}.k8s.{{ $v.BaseDomain }}.'
      HostedZoneId: !Ref 'HostedZone'
      Type: A
  EtcdRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      Name: 'etcd.{{ $v.ClusterID }}.k8s.{{ $v.BaseDomain }}.'
      HostedZoneId: !Ref 'HostedZone'
      TTL: '900'
      Type: CNAME
      ResourceRecords:
        - !GetAtt {{ $v.MasterInstanceResourceName }}.PrivateDnsName
  IngressRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      AliasTarget:
        DNSName: !GetAtt IngressLoadBalancer.DNSName
        HostedZoneId: !GetAtt IngressLoadBalancer.CanonicalHostedZoneNameID
        EvaluateTargetHealth: false
      Name: 'ingress.{{ $v.ClusterID }}.k8s.{{ $v.BaseDomain }}.'
      HostedZoneId: !Ref 'HostedZone'
      Type: A
  IngressWildcardRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      Name: '*.{{ $v.ClusterID }}.k8s.{{ $v.BaseDomain }}.'
      HostedZoneId: !Ref 'HostedZone'
      TTL: '900'
      Type: CNAME
      ResourceRecords:
        - !Ref 'IngressRecordSet'
{{end}}
{{end}}`
