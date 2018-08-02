package guest

const SecurityGroups = `{{define "security_groups" }}
{{- $v := .Guest.SecurityGroups }}
  {{ range $i, $elem :=  $v.MasterSecurityGroups }}
  MasterSecurityGroup{{ $i }}:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: {{ $elem.SecurityGroupName }}
      VpcId: !Ref VPC
      SecurityGroupIngress:
      {{ range $elem.SecurityGroupRules }}
      -
        IpProtocol: {{ .Protocol }}
        FromPort: {{ .Port }}
        ToPort: {{ .Port }}
        CidrIp: {{ .SourceCIDR }}
      {{ end }}
      {{- if $v.APIWhitelistEnabled }}
      -
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: !Join [ "/", [ !Ref NATEIP, "32" ] ]
      {{- end }}
      Tags:
        - Key: Name
          Value:  {{ $elem.SecurityGroupName }}
  {{ end }}
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

  # Allow all access between masters and workers for calico. This is done after
  # the other rules to avoid circular dependencies.
  MasterAllowCalicoIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: MasterSecurityGroup
    Properties:
      GroupId: !Ref MasterSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref MasterSecurityGroup

  MasterAllowWorkerCalicoIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: MasterSecurityGroup
    Properties:
      GroupId: !Ref MasterSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref WorkerSecurityGroup

  WorkerAllowCalicoIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: WorkerSecurityGroup
    Properties:
      GroupId: !Ref WorkerSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref WorkerSecurityGroup

  WorkerAllowMasterCalicoIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: WorkerSecurityGroup
    Properties:
      GroupId: !Ref WorkerSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref MasterSecurityGroup

{{ end }}`
