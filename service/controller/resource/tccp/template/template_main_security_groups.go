package template

const TemplateMainSecurityGroups = `
{{- define "security_groups" -}}
{{- $v := .SecurityGroups -}}
  EtcdPeerSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: {{ $v.ClusterID }}-etcd-peer
      VpcId: !Ref VPC
      Tags:
      - Key: giantswarm.io/security-group-type
        Value: etcd-peer
      SecurityGroupIngress:
      - Description: "Allow traffic for ETCD peers."
        IpProtocol: tcp
        FromPort: 2380
        ToPort: 2380
        CidrIp: {{ $v.TenantClusterVPCCIDR }}
  MasterSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: {{ $v.ClusterID }}-master
      VpcId: !Ref VPC
      SecurityGroupIngress:

      {{- if $v.APIWhitelist.Public.Enabled }}
      #
      # Public API Whitelist Enabled Rules
      #
      -
        Description: "Allow traffic from Control Plane CIDR."
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: {{ $v.ControlPlaneVPCCIDR }}
      -
        Description: "Allow traffic from Tenant Cluster CIDR."
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: {{ $v.TenantClusterVPCCIDR }}

      {{- range $subnet := $v.APIWhitelist.Public.SubnetList }}
      -
        Description: "Custom Public API Whitelist CIDR."
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: {{ $subnet }}
      {{- end }}

      {{- range $v.ControlPlaneNATGatewayAddresses }}
      -
        Description: "Allow traffic from NAT Gateways."
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: {{ .PublicIp }}/32
      {{- end }}

      {{- range .NATGateway.Gateways }}
      -
        Description: "Allow NAT gateway IP."
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: !Join [ "/", [ !Ref {{ .NATEIPName }}, "32" ] ]
      {{- end }}

      {{- else }}
      #
      # Public API Whitelist Disabled Rules
      #
      -
        Description: "Allow all traffic to the master instance."
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: 0.0.0.0/0
      {{- end }}

      -
        Description: "Allow traffic from Control Plane CIDR to 4194 for cadvisor scraping."
        IpProtocol: tcp
        FromPort: 4194
        ToPort: 4194
        CidrIp: {{ $v.ControlPlaneVPCCIDR }}
      -
        Description: "Allow traffic from Control Plane CIDR to 2379 for etcd backup."
        IpProtocol: tcp
        FromPort: 2379
        ToPort: 2379
        CidrIp: {{ $v.ControlPlaneVPCCIDR }}
      -
        Description: "Allow traffic from Control Plane CIDR to 10250 for kubelet scraping."
        IpProtocol: tcp
        FromPort: 10250
        ToPort: 10250
        CidrIp: {{ $v.ControlPlaneVPCCIDR }}
      -
        Description: "Allow traffic from Control Plane CIDR to 10300 for node-exporter scraping."
        IpProtocol: tcp
        FromPort: 10300
        ToPort: 10300
        CidrIp: {{ $v.ControlPlaneVPCCIDR }}
      -
        Description: "Allow traffic from Control Plane CIDR to 10301 for kube-state-metrics scraping."
        IpProtocol: tcp
        FromPort: 10301
        ToPort: 10301
        CidrIp: {{ $v.ControlPlaneVPCCIDR }}
      -
        Description: "Only allow SSH traffic from the Control Plane."
        IpProtocol: tcp
        FromPort: 22
        ToPort: 22
        CidrIp: {{ $v.ControlPlaneVPCCIDR }}

      Tags:
        - Key: Name
          Value: {{ $v.ClusterID }}-master
  EtcdELBSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: {{ $v.ClusterID }}-etcd-elb
      VpcId: !Ref VPC
      SecurityGroupIngress:
      -
        Description: "Allow all Etcd traffic from the VPC to the Etcd load balancer."
        IpProtocol: tcp
        FromPort: 2379
        ToPort: 2379
        CidrIp: 0.0.0.0/0
      -
        Description: "Allow traffic from Control Plane to Etcd port for backup and metrics."
        IpProtocol: tcp
        FromPort: 2379
        ToPort: 2379
        CidrIp: {{ $v.ControlPlaneVPCCIDR }}
      Tags:
        - Key: Name
          Value: {{ $v.ClusterID }}-etcd-elb
  APIInternalELBSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: {{ $v.ClusterID }}-internal-api
      VpcId: !Ref VPC
      SecurityGroupIngress:

      {{- if $v.APIWhitelist.Private.Enabled }}
      #
      # Private API Whitelist Enabled Rules
      #
      -
        Description: "Allow traffic from Control Plane CIDR."
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: {{ $v.ControlPlaneVPCCIDR }}
      -
        Description: "Allow traffic from Tenant Cluster CIDR."
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: {{ $v.TenantClusterVPCCIDR }}

      -
        Description: "Allow traffic from Tenant Cluster CNI CIDR."
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: {{ $v.TenantClusterCNICIDR }}

      {{- range $subnet := $v.APIWhitelist.Private.SubnetList }}
      -
        Description: "Custom Private API Whitelist CIDR."
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: {{ $subnet }}
      {{- end }}

      {{- else }}
      #
      # Private API Whitelist Disabled Rules
      #
      -
        Description: "Allow all traffic to the master instance from A class network."
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: "10.0.0.0/8"
      -
        Description: "Allow all traffic to the master instance from B class network."
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: "172.16.0.0/12"
      -
        Description: "Allow all traffic to the master instance from C class network."
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: "192.168.0.0/16"
      -
        Description: "Allow all traffic to the master instance from CNI (non RFC-1918)."
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: "100.64.0.0/10"
      -
        Description: "Allow all traffic to the master instance from CNI (non RFC-1918)."
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: "198.19.0.0/16"
      {{- end }}

      Tags:
        - Key: Name
          Value: {{ $v.ClusterID }}-internal-api
  AWSCNISecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: "AWS CNI Security Group configured to the ENIConfig CRD."
      VpcId: !Ref VPC
      Tags:
        - Key: Name
          Value: {{ $v.ClusterID }}-aws-cni
  PodsIngressRuleFromMAsters:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: AWSCNISecurityGroup
    Properties:
      Description: Allow traffic from masters to pods.
      GroupId: !Ref AWSCNISecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref MasterSecurityGroup
  PodsAllowPodsCNIIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: AWSCNISecurityGroup
    Properties:
      Description: Allow traffic from pod to pod.
      GroupId: !Ref AWSCNISecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref AWSCNISecurityGroup
  MasterAllowCalicoIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: MasterSecurityGroup
    Properties:
      GroupId: !Ref MasterSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref MasterSecurityGroup
  MasterAllowAPIInternalELBHealthCheck:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn:
      - MasterSecurityGroup
      - APIInternalELBSecurityGroup
    Properties:
      GroupId: !Ref MasterSecurityGroup
      IpProtocol: "tcp"
      FromPort: 8089
      ToPort: 8089
      SourceSecurityGroupId: !Ref APIInternalELBSecurityGroup
  MasterAllowPodsCNIIngressRule:
      Type: AWS::EC2::SecurityGroupIngress
      DependsOn: MasterSecurityGroup
      Properties:
        Description: Allow traffic from pod to master.
        GroupId: !Ref MasterSecurityGroup
        IpProtocol: -1
        FromPort: -1
        ToPort: -1
        SourceSecurityGroupId: !Ref AWSCNISecurityGroup
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
