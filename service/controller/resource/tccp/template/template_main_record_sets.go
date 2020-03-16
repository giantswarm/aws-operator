package template

const TemplateMainRecordSets = `
{{- define "record_sets" -}}
{{- $v := .RecordSets }}
{{- if $v.Route53Enabled -}}
  HostedZone:
    Type: 'AWS::Route53::HostedZone'
    Properties:
      Name: '{{ $v.ClusterID }}.k8s.{{ $v.BaseDomain }}.'
  InternalHostedZone:
    Type: 'AWS::Route53::HostedZone'
    Properties:
      Name: '{{ $v.ClusterID }}.k8s.{{ $v.BaseDomain }}.'
      HostedZoneConfig:
        Comment: "Internal hosted zone for internal network"
      VPCs:
        - VPCId: !Ref VPC
          VPCRegion: '{{ $v.VPCRegion }}'
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
  ApiPublicInternalRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      AliasTarget:
        DNSName: !GetAtt ApiInternalLoadBalancer.DNSName
        HostedZoneId: !GetAtt ApiInternalLoadBalancer.CanonicalHostedZoneNameID
        EvaluateTargetHealth: false
      Name: 'internal-api.{{ $v.ClusterID }}.k8s.{{ $v.BaseDomain }}.'
      HostedZoneId: !Ref 'HostedZone'
      Type: A
  ApiPrivateInternalRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      AliasTarget:
        DNSName: !GetAtt ApiInternalLoadBalancer.DNSName
        HostedZoneId: !GetAtt ApiInternalLoadBalancer.CanonicalHostedZoneNameID
        EvaluateTargetHealth: false
      Name: 'api.{{ $v.ClusterID }}.k8s.{{ $v.BaseDomain }}.'
      HostedZoneId: !Ref 'InternalHostedZone'
      Type: A
  EtcdInternalRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      AliasTarget:
        DNSName: !GetAtt EtcdLoadBalancer.DNSName
        HostedZoneId: !GetAtt EtcdLoadBalancer.CanonicalHostedZoneNameID
        EvaluateTargetHealth: false
      Name: '{{ $v.EtcdDomain }}.'
      HostedZoneId: !Ref 'InternalHostedZone'
      Type: A
  EtcdRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      AliasTarget:
        DNSName: !GetAtt EtcdLoadBalancer.DNSName
        HostedZoneId: !GetAtt EtcdLoadBalancer.CanonicalHostedZoneNameID
        EvaluateTargetHealth: false
      Name: '{{ $v.EtcdDomain }}.'
      HostedZoneId: !Ref 'HostedZone'
      Type: A
  IngressWildcardRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      Name: '*.{{ $v.ClusterID }}.k8s.{{ $v.BaseDomain }}.'
      HostedZoneId: !Ref 'HostedZone'
      TTL: '300'
      Type: CNAME
      ResourceRecords:
        - 'ingress.{{ $v.ClusterID }}.k8s.{{ $v.BaseDomain }}.'
  IngressWildcardInternalRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      Name: '*.{{ $v.ClusterID }}.k8s.{{ $v.BaseDomain }}.'
      HostedZoneId: !Ref 'InternalHostedZone'
      TTL: '300'
      Type: CNAME
      ResourceRecords:
        - 'ingress.{{ $v.ClusterID }}.k8s.{{ $v.BaseDomain }}.'
{{- end -}}
{{- end -}}
`
