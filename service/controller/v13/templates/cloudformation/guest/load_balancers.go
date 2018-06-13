package guest

const LoadBalancers = `{{define "load_balancers"}}
  ApiLoadBalancer:
    Type: AWS::ElasticLoadBalancing::LoadBalancer
    Properties:
      ConnectionSettings:
        IdleTimeout: 1200
      HealthCheck:
        HealthyThreshold: {{ .ELBHealthCheckHealthyThreshold }}
        Interval: {{ .ELBHealthCheckInterval }}
        Target: {{ .APIElbHealthCheckTarget }}
        Timeout: {{ .ELBHealthCheckTimeout }}
        UnhealthyThreshold: {{ .ELBHealthCheckUnhealthyThreshold }}
      Instances:
      - !Ref {{ .MasterInstanceResourceName }}
      Listeners:
      {{ range .APIElbPortsToOpen}}
      - InstancePort: {{ .PortInstance }}
        InstanceProtocol: TCP
        LoadBalancerPort: {{ .PortELB }}
        Protocol: TCP
      {{ end }}
      LoadBalancerName: {{ .APIElbName }}
      Scheme: {{ .APIElbScheme }}
      SecurityGroups:
        - !Ref MasterSecurityGroup
      Subnets:
        - !Ref PublicSubnet

  IngressLoadBalancer:
    Type: AWS::ElasticLoadBalancing::LoadBalancer
    DependsOn: VPCGatewayAttachment
    Properties:
      ConnectionSettings:
        IdleTimeout: 60
      HealthCheck:
        HealthyThreshold: {{ .ELBHealthCheckHealthyThreshold }}
        Interval: {{ .ELBHealthCheckInterval }}
        Target: {{ .IngressElbHealthCheckTarget }}
        Timeout: {{ .ELBHealthCheckTimeout }}
        UnhealthyThreshold: {{ .ELBHealthCheckUnhealthyThreshold }}
      Listeners:
      {{ range .IngressElbPortsToOpen}}
      - InstancePort: {{ .PortInstance }}
        InstanceProtocol: TCP
        LoadBalancerPort: {{ .PortELB }}
        Protocol: TCP
      {{ end }}
      LoadBalancerName: {{ .IngressElbName }}
      Policies:
      - PolicyName: "EnableProxyProtocol"
        PolicyType: "ProxyProtocolPolicyType"
        Attributes:
        - Name: "ProxyProtocol"
          Value: "true"
        InstancePorts:
        {{ range .IngressElbPortsToOpen}}
        - {{ .PortInstance }}
        {{ end }}
      Scheme: {{ .IngressElbScheme }}
      SecurityGroups:
        - !Ref IngressSecurityGroup
      Subnets:
        - !Ref PublicSubnet
{{end}}`
