package template

const TemplateMainSecurityGroups = `
{{- define "security_groups" -}}
{{- $v := .SecurityGroups -}}
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
      {{- $g := .NATGateway }}
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
          Value: {{ $v.MasterSecurityGroupName }}
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
  APIInternalELBSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: {{ $v.APIInternalELBSecurityGroupName }}
      VpcId: !Ref VPC
      SecurityGroupIngress:
      {{ range $v.APIInternalELBSecurityGroupRules }}
      -
        Description: {{ .Description }}
        IpProtocol: {{ .Protocol }}
        FromPort: {{ .Port }}
        ToPort: {{ .Port }}
        CidrIp: {{ .SourceCIDR }}
      {{ end }}
      Tags:
        - Key: Name
          Value: {{ $v.APIInternalELBSecurityGroupName }}
  MasterAllowCalicoIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: MasterSecurityGroup
    Properties:
      GroupId: !Ref MasterSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref MasterSecurityGroup
  MasterAllowEtcdIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: MasterSecurityGroup
    Properties:
      GroupId: !Ref MasterSecurityGroup
      IpProtocol: "tcp"
      FromPort: 2379
      ToPort: 2379
      SourceSecurityGroupId: !Ref EtcdELBSecurityGroup
  VPCDefaultSecurityGroupEgress:
    Type: AWS::EC2::SecurityGroupEgress
    Properties:
      Description: Allow outbound traffic from loopback address.
      GroupId: !GetAtt VPC.DefaultSecurityGroup
      IpProtocol: -1
      CidrIp: 127.0.0.1/32
{{- end -}}
`
