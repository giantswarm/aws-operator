package guest

const RecordSets = `{{define "recordsets"}}
  ApiRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      AliasTarget:
        DNSName: !GetAtt ApiLoadBalancer.DNSName
        HostedZoneId: !GetAtt ApiLoadBalancer.CanonicalHostedZoneNameID
        EvaluateTargetHealth: false
      Name: {{.APIELBDomain}}
      HostedZoneId: {{.APIELBHostedZones}}
      Type: A
  EtcdRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      Name: {{.EtcdELBDomain}}
      HostedZoneId: {{.EtcdELBHostedZones}}
      TTL: '900'
      Type: CNAME
      ResourceRecords:
        - !GetAtt {{ .MasterInstanceID }}.PrivateDnsName
  IngressRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      AliasTarget:
        DNSName: !GetAtt IngressLoadBalancer.DNSName
        HostedZoneId: !GetAtt IngressLoadBalancer.CanonicalHostedZoneNameID
        EvaluateTargetHealth: false
      Name: {{.IngressELBDomain}}
      HostedZoneId: {{.IngressELBHostedZones}}
      Type: A
  IngressWildcardRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      Name: '{{.IngressWildcardELBDomain}}'
      HostedZoneId: {{.IngressELBHostedZones}}
      TTL: '900'
      Type: CNAME
      ResourceRecords:
        - {{.IngressELBDomain}}
{{end}}`
