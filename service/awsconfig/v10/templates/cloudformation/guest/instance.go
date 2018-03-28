package guest

const Instance = `{{define "instance"}}
  {{ .Instance.Master.Instance.ResourceName }}:
    Type: "AWS::EC2::Instance"
    Description: Master instance
    Properties:
      AvailabilityZone: {{ .Instance.Master.AZ }}
      IamInstanceProfile: !Ref MasterInstanceProfile
      ImageId: {{ .Instance.Image.ID }}
      InstanceType: {{ .Instance.Master.Instance.Type }}
      SecurityGroupIds:
      - !Ref MasterSecurityGroup
      SubnetId: !Ref PrivateSubnet
      UserData: {{ .Instance.Master.CloudConfig }}
      Tags:
      - Key: Name
        Value: {{ .Instance.Cluster.ID }}-master
    Metadata:
      AWS::CloudFormation::Init:
        config:
          packages:
            rpm:
              aws-cfn-resource-bridge: https://s3.amazonaws.com/cloudformation-examples/aws-cfn-resource-bridge-0.1-4.noarch.rpm
          files:
            "/etc/cfn/bridge.d/mount.conf":
              content:
                Fn::Join:
                - ''
                - - "[mount]\n"
                  - 'resource_type=Custom::VolumeMount'
                  - queue_url=
                  - Fn::GetAtt:
                    - CustomResourcePipeline
                    - Outputs.CustomResourceQueueURL
                  - "\n"
                  - 'timeout=600'
                  - 'create_action=/home/ec2-user/create.sh'
                  - 'update_action=/home/ec2-user/update.sh'
                  - 'delete_action=/home/ec2-user/delete.sh'
            "/home/ec2-user/create.sh":
              source: https://raw.github.com/awslabs/aws-cfn-custom-resource-examples/master/examples/mount/impl/create.sh
              mode: '000755'
              owner: ec2-user
            "/home/ec2-user/update.sh":
              source: https://raw.github.com/awslabs/aws-cfn-custom-resource-examples/master/examples/mount/impl/update.sh
              mode: '000755'
              owner: ec2-user
            "/home/ec2-user/delete.sh":
              source: https://raw.github.com/awslabs/aws-cfn-custom-resource-examples/master/examples/mount/impl/delete.sh
              mode: '000755'
              owner: ec2-user
          services:
            sysvinit:
              cfn-resource-bridge:
                enabled: 'true'
                ensureRunning: 'true'
                files:
                - "/etc/cfn/bridge.d/mount.conf"
                - "/home/ec2-user/create.sh"
                - "/home/ec2-user/update.sh"
                - "/home/ec2-user/delete.sh"

  EtcdVolume:
    Type: AWS::EC2::Volume
    DependsOn:
    - {{ .Instance.Master.Instance.ResourceName }}
    Properties:
      Encrypted: true
      Size: 100
      VolumeType: gp2
      AvailabilityZone: !GetAtt {{ .Instance.Master.Instance.ResourceName }}.AvailabilityZone
      Tags:
      - Key: Name
        Value: {{ .Instance.Cluster.ID }}-etcd
  MountPoint:
    Type: AWS::EC2::VolumeAttachment
    Properties:
      InstanceId: !Ref {{ .Instance.Master.Instance.ResourceName }}
      VolumeId: !Ref EtcdVolume
      Device: /dev/sdh

  CustomResourcePipeline:
    Type: AWS::CloudFormation::Stack
    Properties:
      TemplateURL: https://s3.amazonaws.com/cloudformation-examples/cr-backend-substack-template.template
  EtcdVolumeMount:
    Type: Custom::VolumeMount
    Version: '1.0'
    DependsOn:
    - MountPoint
    Properties:
      ServiceToken:
        Fn::GetAtt:
        - CustomResourcePipeline
        - Outputs.CustomResourceTopicARN
      Device: "/dev/xvdh"
      MountPoint: "/mnt/analysis"
      FsType: ext3
      Format: 'true'

{{end}}`
