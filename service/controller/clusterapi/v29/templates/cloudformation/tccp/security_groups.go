package tccp

const SecurityGroups = `
{{define "security_groups" }}
{{- $v := .Guest.SecurityGroups }}
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
      {{ end }}
      {{- if $v.APIWhitelistEnabled }}
      {{- $g := .Guest.NATGateway }}
      {{- range $g.Gateways }}
      -
        Description: Allow NAT gateway IP
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: !Join [ "/", [ !Ref {{ .NATEIPName }}, "32" ] ]
      {{- end}}
      {{- end }}
      Tags:
        - Key: Name
          Value:  {{ $v.MasterSecurityGroupName }}

  WorkerSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: {{ $v.WorkerSecurityGroupName }}
      VpcId: !Ref VPC
      SecurityGroupIngress:
      {{ range $v.WorkerSecurityGroupRules }}
      -
        IpProtocol: {{ .Protocol }}
        FromPort: {{ .Port }}
        ToPort: {{ .Port }}
        {{ if .SourceCIDR }}
        CidrIp: {{ .SourceCIDR }}
        {{ else }}
        SourceSecurityGroupId: !Ref {{ .SourceSecurityGroup }}
        {{ end }}
      {{ end }}
      Tags:
        - Key: Name
          Value:  {{ $v.WorkerSecurityGroupName }}

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

  # Allow all access between masters and workers for calico. This is done after
  # the other rules to avoid circular dependencies.
  MasterAllowCalicoIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: MasterSecurityGroup
    Properties:
      # Allow access between masters and workers for calico.
      GroupId: !Ref MasterSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref MasterSecurityGroup

  MasterAllowWorkerCalicoIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: MasterSecurityGroup
    Properties:
      # Allow access between masters and workers for calico.
      GroupId: !Ref MasterSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref WorkerSecurityGroup

  MasterAllowEtcdIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: MasterSecurityGroup
    Properties:
      # Allow access between masters and workers for calico.
      GroupId: !Ref MasterSecurityGroup
      IpProtocol: "tcp"
      FromPort: 2379
      ToPort: 2379
      SourceSecurityGroupId: !Ref EtcdELBSecurityGroup

  WorkerAllowCalicoIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: WorkerSecurityGroup
    Properties:
      # Allow access between masters and workers for calico.
      GroupId: !Ref WorkerSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref WorkerSecurityGroup

  WorkerAllowMasterCalicoIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: WorkerSecurityGroup
    Properties:
      # Allow access between masters and workers for calico.
      GroupId: !Ref WorkerSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref MasterSecurityGroup

  VPCDefaultSecurityGroupEgress:
    Type: AWS::EC2::SecurityGroupEgress
    Properties:
      GroupId: !GetAtt VPC.DefaultSecurityGroup
      Description: "Allow outbound traffic from loopback address."
      IpProtocol: -1
      CidrIp: 127.0.0.1/32
{{ end }}
`
