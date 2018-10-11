package guest

const Instance = `{{ define "instance" }}
{{- $v := .Guest.Instance }}
  {{ $v.Master.Instance.ResourceName }}:
    Type: "AWS::EC2::Instance"
    Description: Master instance
    DependsOn:
    - {{ $v.Master.DockerVolume.ResourceName }}
    - EtcdVolume
    Properties:
      AvailabilityZone: {{ $v.Master.AZ }}
      IamInstanceProfile: !Ref MasterInstanceProfile
      ImageId: {{ $v.Image.ID }}
      InstanceType: {{ $v.Master.Instance.Type }}
      Monitoring: {{ $v.Master.Instance.Monitoring }}
      SecurityGroupIds:
      - !Ref MasterSecurityGroup
      SubnetId: !Ref PrivateSubnet
      UserData: {{ $v.Master.CloudConfig }}
      Tags:
      - Key: Name
        Value: {{ $v.Cluster.ID }}-master
  {{ $v.Master.DockerVolume.ResourceName }}:
    Type: AWS::EC2::Volume
    Properties:
{{ if eq $v.Master.EncrypterBackend "kms" }}
      Encrypted: true
{{ end }}
      Size: 50
      VolumeType: gp2
      AvailabilityZone: {{ $v.Master.AZ }}
      Tags:
      - Key: Name
        Value: {{ $v.Master.DockerVolume.Name }}
  EtcdVolume:
    Type: AWS::EC2::Volume
    Properties:
{{ if eq $v.Master.EncrypterBackend "kms" }}
      Encrypted: true
{{ end }}
      Size: 100
      VolumeType: gp2
      AvailabilityZone: {{ $v.Master.AZ }}
      Tags:
      - Key: Name
        Value: {{ $v.Master.EtcdVolume.Name }}
  {{ $v.Master.Instance.ResourceName }}DockerMountPoint:
    Type: AWS::EC2::VolumeAttachment
    Properties:
      InstanceId: !Ref {{ $v.Master.Instance.ResourceName }}
      VolumeId: !Ref {{ $v.Master.DockerVolume.ResourceName }}
      Device: /dev/xvdc
  {{ $v.Master.Instance.ResourceName }}EtcdMountPoint:
    Type: AWS::EC2::VolumeAttachment
    Properties:
      InstanceId: !Ref {{ $v.Master.Instance.ResourceName }}
      VolumeId: !Ref EtcdVolume
      Device: /dev/xvdh
{{ end }}`
