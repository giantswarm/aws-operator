package template

const TemplateMainSecurityGroups = `
{{define "security_groups" }}
  IngressSecurityGroup:
    Type: AWS::EC2::SecurityGroups
    Properties:
      GroupDescription: {{ .SecurityGroups.IngressSecurityGroupName }}
      VpcId: !Ref VPC
      SecurityGroupIngress:
      {{ range .SecurityGroups.IngressSecurityGroupRules }}
      -
        IpProtocol: {{ .Protocol }}
        FromPort: {{ .Port }}
        ToPort: {{ .Port }}
        CidrIp: {{ .SourceCIDR }}
      {{ end }}
      Tags:
        - Key: Name
          Value: {{ .SecurityGroups.IngressSecurityGroupName }}

  VPCDefaultSecurityGroupEgress:
    Type: AWS::EC2::SecurityGroupEgress
    Properties:
      GroupId: !GetAtt VPC.DefaultSecurityGroup
      Description: "Allow outbound traffic from loopback address."
      IpProtocol: -1
      CidrIp: 127.0.0.1/32

  AllowCalicoIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: SecurityGroup
    Properties:
      # Allow access between masters and workers for calico.
      GroupId: !Ref SecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref SecurityGroup

  AllowMasterCalicoIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: SecurityGroup
    Properties:
      # Allow access between masters and workers for calico.
      GroupId: !Ref SecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref MasterSecurityGroup

  SecurityGroup:
    Type: AWS::EC2::SecurityGroups
    Properties:
      GroupDescription: {{ .SecurityGroups.SecurityGroupName }}
      VpcId: !Ref VPC
      SecurityGroupIngress:
      {{ range .SecurityGroups.SecurityGroupRules }}
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
          Value:  {{ .SecurityGroups.SecurityGroupName }}
{{ end }}
`
