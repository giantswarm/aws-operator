package template

const TemplateMainLoadBalancers = `
{{- define "load_balancers" -}}
{{- $v := .Guest.LoadBalancers }}
  ApiInternalLoadBalancer:
    Type: AWS::ElasticLoadBalancingV2::LoadBalancer
    DependsOn:
      - VPCGatewayAttachment
    Properties:
      Name: {{ $v.APIInternalElbName }}
      Scheme: {{ $v.APIInternalElbScheme }}
      Subnets:
      {{- range $s := $v.PrivateSubnets }}
        - !Ref {{ $s }}
      {{- end}}
      Type: network
  {{- range $v.APIInternalElbListenersAndTargets}}
  {{ .TargetResourceName }}:
    Type: AWS::ElasticLoadBalancingV2::TargetGroup
    Properties:
      HealthCheckEnabled: true
      HealthCheckIntervalSeconds: {{ $v.ELBHealthCheckInterval }}
      HealthCheckPort: {{ .PortInstance }}
      HealthCheckProtocol: TCP
      HealthyThresholdCount: {{ $v.ELBHealthCheckHealthyThreshold }}
      Port: {{ .PortInstance }}
      Protocol: TCP
      Targets:
      - Id: !Ref {{ $v.MasterInstanceResourceName }}
      TargetType: instance
      UnhealthyThresholdCount: {{ $v.ELBHealthCheckUnhealthyThreshold }}
      VpcId: !Ref VPC
  {{ .ListenerResourceName }}:
    Type: AWS::ElasticLoadBalancingV2::Listener
    Properties:
      DefaultActions:
      - Type: forward
        TargetGroupArn: !Ref {{ .TargetResourceName }}
      LoadBalancerArn: !Ref ApiInternalLoadBalancer
      Port: {{ .PortELB }}
      Protocol: TCP
  {{- end }}
  ApiLoadBalancer:
    Type: AWS::ElasticLoadBalancingV2::LoadBalancer
    DependsOn:
      - VPCGatewayAttachment
    Properties:      
      Name: {{ $v.APIElbName }}
      Scheme: {{ $v.APIElbScheme }}  
      Subnets:
      {{- range $s := $v.PublicSubnets }}
        - !Ref {{ $s }}
      {{end}}
      Type: network
  {{- range $v.APIElbListenersAndTargets}}
  {{ .TargetResourceName }}:
    Type: AWS::ElasticLoadBalancingV2::TargetGroup
    Properties:
      HealthCheckEnabled: true
      HealthCheckIntervalSeconds: {{ $v.ELBHealthCheckInterval }}
      HealthCheckPort: {{ .PortInstance }}
      HealthCheckProtocol: TCP
      HealthyThresholdCount: {{ $v.ELBHealthCheckHealthyThreshold }}
      Port: {{ .PortInstance }}
      Protocol: TCP
      Targets:
      - Id: !Ref {{ $v.MasterInstanceResourceName }}
      TargetType: instance
      UnhealthyThresholdCount: {{ $v.ELBHealthCheckUnhealthyThreshold }}
      VpcId: !Ref VPC
  {{ .ListenerResourceName }}:
    Type: AWS::ElasticLoadBalancingV2::Listener
    Properties:
      DefaultActions:
      - Type: forward
        TargetGroupArn: !Ref {{ .TargetResourceName }}
      LoadBalancerArn: !Ref ApiLoadBalancer
      Port: {{ .PortELB }}
      Protocol: TCP
  {{- end }}
  EtcdLoadBalancer:
    Type: AWS::ElasticLoadBalancingV2::LoadBalancer
    Properties:      
      Name: {{ $v.EtcdElbName }}
      Scheme: {{ $v.EtcdElbScheme }}
      Subnets:
      {{- range $s := $v.PrivateSubnets }}
        - !Ref {{ $s }}
      {{end}}
      Type: network
  {{- range $v.EtcdElbListenersAndTargets}}
  {{ .TargetResourceName }}:
    Type: AWS::ElasticLoadBalancingV2::TargetGroup
    Properties:
      HealthCheckEnabled: true
      HealthCheckIntervalSeconds: {{ $v.ELBHealthCheckInterval }}
      HealthCheckPort: {{ .PortInstance }}
      HealthCheckProtocol: TCP
      HealthyThresholdCount: {{ $v.ELBHealthCheckHealthyThreshold }}
      Port: {{ .PortInstance }}
      Protocol: TCP
      Targets:
      - Id: !Ref {{ $v.MasterInstanceResourceName }}
      TargetType: instance
      UnhealthyThresholdCount: {{ $v.ELBHealthCheckUnhealthyThreshold }}
      VpcId: !Ref VPC
  {{ .ListenerResourceName }}:
    Type: AWS::ElasticLoadBalancingV2::Listener
    Properties:
      DefaultActions:
      - Type: forward
        TargetGroupArn: !Ref {{ .TargetResourceName }}
      LoadBalancerArn: !Ref EtcdLoadBalancer
      Port: {{ .PortELB }}
      Protocol: TCP
  {{- end }}
  IngressLoadBalancer:
    Type: AWS::ElasticLoadBalancingV2::LoadBalancer
    DependsOn:
      - VPCGatewayAttachment
    Properties:      
      Name: {{ $v.IngressElbName }}   
      Scheme: {{ $v.IngressElbScheme }}
      Subnets:
      {{- range $s := $v.PublicSubnets }}
        - !Ref {{ $s }}
      {{end}}
      Type: network
  {{- range $v.IngressElbListenersAndTargets}}
  {{ .TargetResourceName }}:
    Type: AWS::ElasticLoadBalancingV2::TargetGroup
    Properties:
      HealthCheckEnabled: true
      HealthCheckIntervalSeconds: {{ $v.ELBHealthCheckInterval }}
      HealthCheckPort: {{ .PortInstance }}
      HealthCheckProtocol: TCP
      HealthyThresholdCount: {{ $v.ELBHealthCheckHealthyThreshold }}
      Port: {{ .PortInstance }}
      Protocol: TCP
      TargetGroupAttributes:
      - proxy_protocol_v2.enabled: true
      TargetType: instance
      UnhealthyThresholdCount: {{ $v.ELBHealthCheckUnhealthyThreshold }}
      VpcId: !Ref VPC
  {{ .ListenerResourceName }}:
    Type: AWS::ElasticLoadBalancingV2::Listener
    Properties:
      DefaultActions:
      - Type: forward
        TargetGroupArn: !Ref {{ .TargetResourceName }}
      LoadBalancerArn: !Ref IngressLoadBalancer
      Port: {{ .PortELB }}
      Protocol: TCP
  {{- end }}
{{- end -}}
`
