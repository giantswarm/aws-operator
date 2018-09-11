package setup

var encrypterVaultTemplate = `AWSTemplateFormatVersion: 2010-09-09
Description: E2E encrypter Vault.

Parameters:
  AccessKey:
    Type: String
  SecretKeyId:
    Type: String

Resources:
  VaultSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupName: "Internet Group"
      GroupDescription: "SSH traffic in, all traffic out."
      SecurityGroupIngress:
      - IpProtocol: tcp
        FromPort: '22'
        ToPort: '22'
        CidrIp: 0.0.0.0/0
      - IpProtocol: tcp
        FromPort: '8200'
        ToPort: '8200'
        CidrIp: 0.0.0.0/0
      Tags:
      - Key: Name
        Value: vault-poc

  VaultInstance:
    Type: AWS::EC2::Instance
    Properties:
      ImageId: ami-4f508c22
      InstanceType: t2.micro
      KeyName: vault-poc
      NetworkInterfaces: 
        - AssociatePublicIpAddress: "true"
          DeleteOnTermination : "true"
          DeviceIndex: "0"
          GroupSet: 
            - Ref: VaultSecurityGroup
      UserData:
        Fn::Base64: !Sub
          - |
            #!/bin/bash
            apt update && apt install -y unzip jq
            wget https://releases.hashicorp.com/vault/0.10.1/vault_0.10.1_linux_amd64.zip
            unzip vault_0.10.1_linux_amd64.zip -d /usr/local/bin
            export VAULT_ADDR=http://127.0.0.1:8200
            cat << 'SCRIPT' > /usr/local/bin/vaultinit.sh
            #!/usr/bin/env bash
            ################
            # Vault generic setup
            ################
            cat <<EOF > config.hcl
            storage "inmem" {}
            listener "tcp" {
              address = "0.0.0.0:8200"
              tls_disable = 1
            }
            disable_mlock = true
            EOF
            vault server -config=config.hcl 2>&1 > /var/log/vault.txt &
            sleep 3
            curl \
              --silent \
              --request PUT \
              --data '{"secret_shares": 1, "secret_threshold": 1}' \
              $VAULT_ADDR/v1/sys/init | tee >(jq -r .root_token > /tmp/root_token) >(jq -r .keys[0] > /tmp/key)
            key=$(cat /tmp/key)
            curl \
              --silent \
              --request PUT \
              --data '{"key": "'"$key"'"}' \
              $VAULT_ADDR/v1/sys/unseal
            export VAULT_TOKEN=$(cat /tmp/root_token)
            ################
            # End of Vault generic setup
            ################
            ################
            # Transist secrets backend
            ################
            vault secrets enable transit
            ################
            # End of transist secrets backend
            ################
            ################
            # Policies
            ################
            cat <<EOF > transit-policy.hcl
            path "transit/*" {
              capabilities = ["create", "read", "update", "delete", "list"]
            }
            EOF
            vault write sys/policy/transit policy=@transit-policy.hcl
            cat <<EOF > auth-aws-role-admin-policy.hcl
            path "auth/aws/role/*" {
              capabilities = ["create", "read", "update", "delete", "list"]
            }
            EOF
            vault write sys/policy/auth-aws-admin policy=@auth-aws-role-admin-policy.hcl
            ################
            # End of Policies
            ################
            ################
            # AWS auth backend
            ################
            vault auth enable aws
            client_cfg_payload=$(cat <<EOF
            {
              "access_key": "${AccessKey}",
              "secret_key": "${SecretKeyId}",
            }
            EOF
            )
            curl \
              --header "X-Vault-Token: $VAULT_TOKEN" \
              --request POST \
              --data "$client_cfg_payload" \
              $VAULT_ADDR/v1/auth/aws/config/client
            ################
            # End of AWS auth backend
            ################
            ################
            # AWS auth roles
            ################
            encrypter_role_payload=$(cat <<EOF
            {
              "auth_type": "ec2",
              "bound_region": "eu-central-1",
              "bound_iam_role_arn": "",
              "policies": "auth-aws-admin,transit",
              "max_ttl": 1800000,
              "disallow_reauthentication": false,
              "allow_instance_migration": false
            }
            EOF
            )
            curl \
              --header "X-Vault-Token: $VAULT_TOKEN" \
              --request POST \
              --data "$encrypter_role_payload" \
              $VAULT_ADDR/v1/auth/aws/role/encrypter
            decrypter_role_payload=$(cat <<EOF
            {
              "auth_type": "ec2",
              "bound_region": "eu-central-1",
              "bound_iam_role_arn": "",
              "policies": "transit",
              "max_ttl": 1800000,
              "disallow_reauthentication": false,
              "allow_instance_migration": false
            }
            EOF
            )
            curl \
              --header "X-Vault-Token: $VAULT_TOKEN" \
              --request POST \
              --data "$decrypter_role_payload" \
              $VAULT_ADDR/v1/auth/aws/role/decrypter
            ################
            # End of AWS auth roles
            ################
            SCRIPT
            chmod a+x /usr/local/bin/vaultinit.sh
            vaultinit.sh

          - {}
      Tags:
      - Key: Name
        Value: vault-poc-vault

Outputs:
  VaultAddress:
    Description: Vault EC2 instance public IP.
    Value: !GetAtt VaultInstance.PublicIp
`
