package guest

const Instances = `{{define "instance"}}
  {{ .Instances.Master.Instance.ID }}:
    Type: "AWS::EC2::Instance"
    Description: Master instance
    Properties:
      AvailabilityZone: {{ .Instances.Master.AZ }}
      IamInstanceProfile: !Ref MasterInstanceProfile
      ImageId: {{ .Instances.Image.ID }}
      InstanceType: {{ .Instances.Master.Instance.Type }}
      SecurityGroupIds:
      - !Ref MasterSecurityGroup
      SubnetId: !Ref PrivateSubnet
      UserData: {{ .Instances.Master.CloudConfig }}
      Tags:
      - Key: Name
        Value: {{ .Instances.Cluster.ID }}-master
  EtcdVolume:
    Type: AWS::EC2::Volume
    DependsOn:
    - {{ .Instances.Master.Instance.ID }}
    Properties:
      Encrypted: true
      Size: 100
      VolumeType: gp2
      AvailabilityZone: !GetAtt {{ .Instances.Master.Instance.ID }}.AvailabilityZone
      Tags:
      - Key: Name
        Value: {{ .Instances.Cluster.ID }}-etcd
  MountPoint:
    Type: AWS::EC2::VolumeAttachment
    Properties:
      InstanceId: !Ref {{ .Instances.Master.Instance.ID }}
      VolumeId: !Ref EtcdVolume
      Device: /dev/sdh
{{end}}`
