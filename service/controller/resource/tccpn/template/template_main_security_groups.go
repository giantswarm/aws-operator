package template

const TemplateMainSecurityGroups = `
{{- define "security_groups" -}}
{{- $v := .SecurityGroups -}}
  MasterSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: {{ $v.MasterSecurityGroupName }}
      VpcId: {{ $v.VPCID }}
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
          Value:  {{ $v.MasterSecurityGroupName }}
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
  EtcdELBSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: {{ $v.EtcdELBSecurityGroupName }}
      VpcId: {{ $v.VPCID }}
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
  MasterAllowAllIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: MasterSecurityGroup
    Properties:
      GroupId: !Ref MasterSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref MasterSecurityGroup
  MasterAllowPodsCNIIngressRule:
      Type: AWS::EC2::SecurityGroupIngress
      DependsOn: MasterSecurityGroup
      Properties:
        Description: Allow traffic from pod to master.
        GroupId: !Ref MasterSecurityGroup
        IpProtocol: -1
        FromPort: -1
        ToPort: -1
        SourceSecurityGroupId: {{ $v.AWSCNISecurityGroupID }}
  MasterAllowEtcdIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: MasterSecurityGroup
    Properties:
      GroupId: !Ref MasterSecurityGroup
      IpProtocol: "tcp"
      FromPort: 2379
      ToPort: 2379
      SourceSecurityGroupId: !Ref EtcdELBSecurityGroup
{{- end -}}
`
