package guest

// TODO rename to DNS
const RecordSets = `{{define "recordsets"}}
{{ if .Route53Enabled }}
  HostedZone: 
    Type: "AWS::Route53::HostedZone"
    Properties: 
      Name: {{.HostedZoneDomain}}
  ApiRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      AliasTarget:
        DNSName: !GetAtt ApiLoadBalancer.DNSName
        HostedZoneId: !GetAtt ApiLoadBalancer.CanonicalHostedZoneNameID
        EvaluateTargetHealth: false
      Name: api.{{.HostedZoneDomain}}
      HostedZoneId: !Ref HostedZone
      Type: A
  EtcdRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      Name: etcd.{{.HostedZoneDomain}}
      HostedZoneId: !Ref HostedZone
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
      Name: ingress.{{.HostedZoneDomain}}
      HostedZoneId: !Ref HostedZone
      Type: A
  IngressWildcardRecordSet:
    Type: AWS::Route53::RecordSet
    Properties:
      Name: *.{{.HostedZoneDomain}}
      HostedZoneId: !Ref HostedZone
      TTL: '900'
      Type: CNAME
      ResourceRecords:
        - {{.IngressELBDomain}}
{{end}}
{{end}}`
