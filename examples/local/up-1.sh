#!/bin/bash -x

. ./env.sh 

for f in *.tmpl.yaml; do
    sed \
        -e 's|${CLUSTER_NAME}|'"${CLUSTER_NAME}"'|g' \
        -e 's|${COMMON_DOMAIN}|'"${COMMON_DOMAIN}"'|g' \
        -e 's|${ID_RSA_PUB}|'"${ID_RSA_PUB}"'|g' \
        -e 's|${AWS_ACCESS_KEY_ID}|'"${AWS_ACCESS_KEY_ID}"'|g' \
        -e 's|${AWS_SECRET_ACCESS_KEY}|'"${AWS_SECRET_ACCESS_KEY}"'|g' \
        -e 's|${AWS_SESSION_TOKEN}|'"${AWS_SESSION_TOKEN}"'|g' \
        -e 's|${AWS_REGION}|'"${AWS_REGION}"'|g' \
        -e 's|${AWS_AZ}|'"${AWS_AZ}"'|g' \
        -e 's|${AWS_AMI}|'"${AWS_AMI}"'|g' \
        -e 's|${AWS_INSTANCE_TYPE_MASTER}|'"${AWS_INSTANCE_TYPE_MASTER}"'|g' \
        -e 's|${AWS_INSTANCE_TYPE_WORKER}|'"${AWS_INSTANCE_TYPE_WORKER}"'|g' \
        -e 's|${AWS_API_HOSTED_ZONE}|'"${AWS_API_HOSTED_ZONE}"'|g' \
        -e 's|${AWS_INGRESS_HOSTED_ZONE}|'"${AWS_INGRESS_HOSTED_ZONE}"'|g' \
        -e 's|${AWS_VPC_PEER_ID}|'"${AWS_VPC_PEER_ID}"'|g' \
        ./$f > ./${f%.tmpl.yaml}.yaml
done

eval $(minikube docker-env)
(
    cd ../..
    GOOS=linux go build github.com/giantswarm/aws-operator
    docker build -t quay.io/giantswarm/aws-operator:local-dev .
)

kubectl apply -f ./configmap.yaml
kubectl apply -f ./configmap-ssh.yaml
kubectl apply -f ./deployment.yaml

sleep 5 

kubectl create -f ./cluster.yaml