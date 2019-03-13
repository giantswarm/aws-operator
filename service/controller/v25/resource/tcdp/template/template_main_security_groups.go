package template

const TemplateMainSecurityGroups = `
{{define "security_groups" }}
  SecurityGroup:
    Type: AWS::EC2::SecurityGroups
    Properties:
      GroupDescription: {{ .SecurityGroups.SecurityGroupName }}
      VpcId: !Ref VPC
      SecurityGroupIngress:

			# Allow traffic from control plane CIDR to 22 for SSH access.
      -
        IpProtocol: tcp
        FromPort: 22
        ToPort: 22
        CidrIp: {{ .SecurityGroups.ControlPlane.VPC.CIDR }}

			# Allow traffic from control plane CIDR to 4194 for cadvisor scraping.
      -
        IpProtocol: tcp
        FromPort: 4194
        ToPort: 4194
        CidrIp: {{ .SecurityGroups.ControlPlane.VPC.CIDR }}

			# Allow traffic from control plane CIDR to 10250 for kubelet scraping.
      -
        IpProtocol: tcp
        FromPort: 10250
        ToPort: 10250
        CidrIp: {{ .SecurityGroups.ControlPlane.VPC.CIDR }}

			# Allow traffic from control plane CIDR to 10300 for node-exporter scraping.
      -
        IpProtocol: tcp
        FromPort: 10300
        ToPort: 10300
        CidrIp: {{ .SecurityGroups.ControlPlane.VPC.CIDR }}

			# Allow traffic from control plane CIDR to 10301 for kube-state-metrics scraping.
      -
        IpProtocol: tcp
        FromPort: 10301
        ToPort: 10301
        CidrIp: {{ .SecurityGroups.ControlPlane.VPC.CIDR }}

      -
        IpProtocol: {{ .Protocol }}
        FromPort: {{ .Port }}
        ToPort: {{ .Port }}
        SourceSecurityGroupId: !Ref {{ .SourceSecurityGroup }}
      Tags:
        - Key: Name
          Value:  {{ .SecurityGroups.SecurityGroupName }}
{{ end }}
`





		{
			Description:         "Allow traffic from the ingress security group to the ingress controller port 443.",
			Port:                key.IngressControllerSecurePort(customObject),
			Protocol:            tcpProtocol,
			SourceSecurityGroup: ingressSecurityGroupName,
		},
		{
			Description:         "Allow traffic from the ingress security group to the ingress controller port 80.",
			Port:                key.IngressControllerInsecurePort(customObject),
			Protocol:            tcpProtocol,
			SourceSecurityGroup: ingressSecurityGroupName,
		},
