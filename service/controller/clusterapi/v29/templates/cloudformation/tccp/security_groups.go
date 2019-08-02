package tccp

const SecurityGroups = `
{{- define "security_groups" -}}
{{- $v := .Guest.SecurityGroups -}}
  MasterSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: {{ $v.MasterSecurityGroupName }}
      VpcId: !Ref VPC
      SecurityGroupIngress:
      {{ range $v.MasterSecurityGroupRules }}
      -
        Description: {{ .Description }}
        IpProtocol: {{ .Protocol }}
        FromPort: {{ .Port }}
        ToPort: {{ .Port }}
        CidrIp: {{ .SourceCIDR }}
      {{- end }}
      {{- if $v.APIWhitelistEnabled }}
      {{- $g := .Guest.NATGateway }}
      {{- range $g.Gateways }}
      -
        Description: Allow NAT gateway IP
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: !Join [ "/", [ !Ref {{ .NATEIPName }}, "32" ] ]
      {{- end }}
      {{- end }}
      Tags:
        - Key: Name
          Value:  {{ $v.MasterSecurityGroupName }}
        - Key: giantswarm.io/tccp
          Value: true
  IngressSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: {{ $v.IngressSecurityGroupName }}
      VpcId: !Ref VPC
      SecurityGroupIngress:
      {{ range $v.IngressSecurityGroupRules }}
      -
        IpProtocol: {{ .Protocol }}
        FromPort: {{ .Port }}
        ToPort: {{ .Port }}
        CidrIp: {{ .SourceCIDR }}
      {{ end }}
      Tags:
        - Key: Name
          Value: {{ $v.IngressSecurityGroupName }}
        - Key: giantswarm.io/tccp
          Value: true
  EtcdELBSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: {{ $v.EtcdELBSecurityGroupName }}
      VpcId: !Ref VPC
      SecurityGroupIngress:
      {{ range $v.EtcdELBSecurityGroupRules }}
      -
        IpProtocol: {{ .Protocol }}
        FromPort: {{ .Port }}
        ToPort: {{ .Port }}
        CidrIp: {{ .SourceCIDR }}
      {{ end }}
      Tags:
        - Key: Name
          Value: {{ $v.EtcdELBSecurityGroupName }}
        - Key: giantswarm.io/tccp
          Value: true
  VPCDefaultSecurityGroupEgress:
    Type: AWS::EC2::SecurityGroupEgress
    Properties:
      GroupId: !GetAtt VPC.DefaultSecurityGroup
      Description: "Allow outbound traffic from loopback address."
      IpProtocol: -1
      CidrIp: 127.0.0.1/32
{{- end -}}
`
