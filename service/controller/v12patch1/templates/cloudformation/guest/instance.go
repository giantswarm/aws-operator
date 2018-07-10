package guest

const Instance = `{{define "instance"}}
  {{ .Instance.Master.Instance.ResourceName }}:
    Type: "AWS::EC2::Instance"
    Description: Master instance
    DependsOn:
    - {{ .Instance.Master.DockerVolume.ResourceName }}
    - EtcdVolume
    Properties:
      AvailabilityZone: {{ .Instance.Master.AZ }}
      IamInstanceProfile: !Ref MasterInstanceProfile
      ImageId: {{ .Instance.Image.ID }}
      InstanceType: {{ .Instance.Master.Instance.Type }}
      Monitoring: {{ .Instance.Master.Instance.Monitoring }}
      SecurityGroupIds:
      - !Ref MasterSecurityGroup
      SubnetId: !Ref PrivateSubnet
      UserData: {{ .Instance.Master.CloudConfig }}
      Tags:
      - Key: Name
        Value: {{ .Instance.Cluster.ID }}-master
  {{ .Instance.Master.DockerVolume.ResourceName }}:
    Type: AWS::EC2::Volume
    Properties:
      Encrypted: true
      Size: 50
      VolumeType: gp2
      AvailabilityZone: {{ .Instance.Master.AZ }}
      Tags:
      - Key: Name
        Value: {{ .Instance.Master.DockerVolume.Name }}
  EtcdVolume:
    Type: AWS::EC2::Volume
    Properties:
      Encrypted: true
      Size: 100
      VolumeType: gp2
      AvailabilityZone: {{ .Instance.Master.AZ }}
      Tags:
      - Key: Name
        Value: {{ .Instance.Master.EtcdVolume.Name }}
  {{ .Instance.Master.Instance.ResourceName }}DockerMountPoint:
    Type: AWS::EC2::VolumeAttachment
    Properties:
      InstanceId: !Ref {{ .Instance.Master.Instance.ResourceName }}
      VolumeId: !Ref {{ .Instance.Master.DockerVolume.ResourceName }}
      Device: /dev/xvdc
  {{ .Instance.Master.Instance.ResourceName }}EtcdMountPoint:
    Type: AWS::EC2::VolumeAttachment
    Properties:
      InstanceId: !Ref {{ .Instance.Master.Instance.ResourceName }}
      VolumeId: !Ref EtcdVolume
      Device: /dev/xvdh
{{end}}`
