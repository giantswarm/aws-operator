package template

const TemplateMainSecurityGroups = `
{{- define "security_groups" -}}
  GeneralSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: General Node Pool Security Group For Basic Traffic Rules.
      SecurityGroupIngress:
      -
        Description: Allow traffic from control plane CIDR to 22 for SSH access.
        IpProtocol: tcp
        FromPort: 22
        ToPort: 22
        CidrIp: {{ .SecurityGroups.ControlPlane.VPC.CIDR }}
      -
        Description: Allow traffic from control plane CIDR to 4194 for cadvisor scraping.
        IpProtocol: tcp
        FromPort: 4194
        ToPort: 4194
        CidrIp: {{ .SecurityGroups.ControlPlane.VPC.CIDR }}
      -
        Description: Allow traffic from control plane CIDR to 10250 for kubelet scraping.
        IpProtocol: tcp
        FromPort: 10250
        ToPort: 10250
        CidrIp: {{ .SecurityGroups.ControlPlane.VPC.CIDR }}
      -
        Description: Allow traffic from control plane CIDR to 10300 for node-exporter scraping.
        IpProtocol: tcp
        FromPort: 10300
        ToPort: 10300
        CidrIp: {{ .SecurityGroups.ControlPlane.VPC.CIDR }}
      -
        Description: Allow traffic from control plane CIDR to 10301 for kube-state-metrics scraping.
        IpProtocol: tcp
        FromPort: 10301
        ToPort: 10301
        CidrIp: {{ .SecurityGroups.ControlPlane.VPC.CIDR }}
      -
        Description: Allow traffic from the ingress security group to 443 for ingress-controller.
        IpProtocol: tcp
        FromPort: 30011
        ToPort: 30011
        SourceSecurityGroupId: {{ .SecurityGroups.TenantCluster.Ingress.ID }}
      -
        Description: Allow traffic from the ingress security group to 80 for ingress-controller.
        IpProtocol: tcp
        FromPort: 30010
        ToPort: 30010
        SourceSecurityGroupId: {{ .SecurityGroups.TenantCluster.Ingress.ID }}
      VpcId: {{ .SecurityGroups.TenantCluster.VPC.ID }}
  GeneralMasterIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: GeneralSecurityGroup
    Properties:
      Description: Allow traffic from the TCNP General Security Group to the TCCP Master Security Group.
      GroupId: {{ .SecurityGroups.TenantCluster.Master.ID }}
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref GeneralSecurityGroup
  InternalIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: GeneralSecurityGroup
    Properties:
      Description: Allow traffic between workloads within the Node Pool.
      GroupId: !Ref GeneralSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref GeneralSecurityGroup
  MasterGeneralIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: GeneralSecurityGroup
    Properties:
      Description: Allow traffic from the TCCP Master Security Group to the TCNP General Security Group.
      GroupId: !Ref GeneralSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: {{ .SecurityGroups.TenantCluster.Master.ID }}
{{- end -}}
`
