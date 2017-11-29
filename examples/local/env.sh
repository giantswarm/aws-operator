#!/bin/bash

if [ -z ${AWS_ACCESS_KEY_ID} ] ; then
  echo "AWS_ACCESS_KEY_ID not set"
  exit 1
fi

if [ -z ${AWS_SECRET_ACCESS_KEY} ] ; then
  echo "AWS_SECRET_ACCESS_KEY not set"
  exit 1
fi

if [ -z ${AWS_SESSION_TOKEN} ] ; then
  echo "AWS_SESSION_TOKEN not set"
  exit 1
fi

if [ -z ${AWS_AMI} ] ; then
  echo "AWS_AMI not set"
  exit 1
fi

if [ -z ${AWS_API_HOSTED_ZONE} ] ; then
  echo "AWS_API_HOSTED_ZONE not set"
  exit 1
fi

if [ -z ${AWS_INGRESS_HOSTED_ZONE} ] ; then
  echo "AWS_INGRESS_HOSTED_ZONE not set"
  exit 1
fi

if [ -z ${AWS_VPC_PEER_ID} ] ; then
  echo "AWS_VPC_PEER_ID not set"
  exit 1
fi

DEFAULT_ID_RSA_PUB=$(cat ~/.ssh/id_rsa.pub)

export CLUSTER_NAME=${CLUSTER_NAME:-g8s}
export COMMON_DOMAIN=${COMMON_DOMAIN:-local}
export ID_RSA_PUB=${ID_RSA_PUB:-"$DEFAULT_ID_RSA_PUB"}
export AWS_REGION=${AWS_REGION:-eu-central-1}
export AWS_AZ=${AWS_AZ:-eu-central-1a}
export AWS_INSTANCE_TYPE_MASTER=${AWS_INSTANCE_TYPE_MASTER:-m3.medium}
export AWS_INSTANCE_TYPE_WORKER=${AWS_INSTANCE_TYPE_WORKER:-m3.medium}