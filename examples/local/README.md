# Running aws-operator Locally

**Note:** This should only be used for testing and development. See the
[/kubernetes/][kubernetes-dir] directory and [Secrets][secrets-doc] for
a production ready configuration.

[kubernetes-dir]: https://github.com/giantswarm/aws-operator/tree/master/kubernetes
[secrets-doc]: https://github.com/giantswarm/aws-operator#secret

This guide explains how to get aws-operator running locally - on minikube, for
example. Also how to create a managed Kubernetes cluster with single master and
single worker using the operator.

All commands are assumed to be run from `examples/local` directory.


## Preparing Templates

All yaml files in this directory are templates. Before continuing with this
guide, all placeholders must be replaced with sensible values.

- *CLUSTER_NAME* - Cluster's name.
- *COMMON_DOMAIN* - Cluster's etcd and API common domain.
- *COMMON_DOMAIN_INTERNAL* - Ingress common domain.
- *ID_RSA_PUB* - SSH public key to be installed on nodes.
- *AWS_ACCESS_KEY_ID* - AWS access key.
- *AWS_SECRET_ACCESS_KEY* - AWS secret.
- *AWS_REGION* - AWS region.
- *AWS_AZ* - AWS availability zone.
- *AWS_AMI* - AWS image to be used on both master and worker machines.
- *AWS_INSTANCE_TYPE_MASTER* - Master machines instance type.
- *AWS_INSTANCE_TYPE_WORKER* - Worker machines instance type.
- *AWS_API_HOSTED_ZONE* - Route 53 hosted zone for API and Etcd
- *AWS_INGRESS_HOSTED_ZONE* - Route 53 hosted zone for Ingress
- *AWS_VPC_PEER_ID* - VPC ID of the host cluster to peer with.

This is a handy snippet that makes it painless - works in bash and zsh.

```bash
export CLUSTER_NAME="example-cluster"
export COMMON_DOMAIN="company.com"
export COMMON_DOMAIN_INTERNAL="internal.company.com"
export ID_RSA_PUB="$(cat ~/.ssh/id_rsa.pub)"
export AWS_ACCESS_KEY_ID="AKIAIXXXXXXXXXXXXXXX"
export AWS_SECRET_ACCESS_KEY="XXXXXXXXXXXXXXXXX/XXXXXXXXXXXXXXXXXXXXXX"
export AWS_REGION="eu-central-1"
export AWS_AZ="eu-central-1a"
export AWS_AMI="ami-d60ad6b9"
export AWS_INSTANCE_TYPE_MASTER="m3.medium"
export AWS_INSTANCE_TYPE_WORKER="m3.medium"
export AWS_API_HOSTED_ZONE="Z*************"
export AWS_INGRESS_HOSTED_ZONE="Z*************"
export AWS_VPC_PEER_ID="vpc-********"

for f in *.tmpl.yaml; do
    sed \
        -e 's|${CLUSTER_NAME}|'"${CLUSTER_NAME}"'|g' \
        -e 's|${COMMON_DOMAIN}|'"${COMMON_DOMAIN}"'|g' \
        -e 's|${COMMON_DOMAIN_INTERNAL}|'"${COMMON_DOMAIN_INTERNAL}"'|g' \
        -e 's|${ID_RSA_PUB}|'"${ID_RSA_PUB}"'|g' \
        -e 's|${AWS_ACCESS_KEY_ID}|'"${AWS_ACCESS_KEY_ID}"'|g' \
        -e 's|${AWS_SECRET_ACCESS_KEY}|'"${AWS_SECRET_ACCESS_KEY}"'|g' \
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
```

- Note: `|` characters are used in `sed` substitution to avoid escaping.


## Cluster Certificates

The easiest way to create certificates is to use the local [cert-operator]
setup. See [this guide][cert-operator-local-setup] for details.

- Note: `CLUSTER_NAME` and `COMMON_DOMAIN` values must match the values used
  during this guide.

## Cluster-Local Docker Image

The operator needs a connection to the K8s API. The simplest approach is to run
the operator as a deployment and use the "in cluster" configuration.

In that case the Docker image needs to be accessible from the K8s cluster
running the operator. For Minikube run `eval $(minikube docker-env)` before
`docker build`, see [reusing the Docker daemon] for details.

[reusing the docker daemon]: https://github.com/kubernetes/minikube/blob/master/docs/reusing_the_docker_daemon.md

```bash
# Optional. Only when using Minikube.
eval $(minikube docker-env)

GOOS=linux go build github.com/giantswarm/aws-operator
docker build -t quay.io/giantswarm/aws-operator:local-dev .

# Optional. Restart running operator after image update.
# Does nothing when the operator is not deployed.
#kubectl delete pod -l app=aws-operator-local
```


## Operator Startup

The operator requires some configuration:

- AWS credentials
- SSH public key to be installed

One way is to provide these with ConfigMaps. Please read the introduction of
this guide if you want to do it more securely.

```bash
kubectl apply -f ./configmap.yaml
kubectl apply -f ./configmap-ssh.yaml
kubectl apply -f ./deployment.yaml
```


## Creating And Connecting New Cluster

First, let's create an new cluster ThirdPartyObject.

```bash
kubectl create -f ./cluster.yaml
```

Creating ThirdPartyObject should eventually result in a working K8s cluster on
AWS. To test if the cluster is ready check the operator's pod logs with the
command below.

```bash
kubectl logs -l app=aws-operator-local
```

When a similar message appears in the log output, the cluster is ready.

```
{"caller":"github.com/giantswarm/aws-operator/service/create/service.go:967","info":"cluster 'pawel' processed","time":"17-05-24 15:24:08.537"}
```

Now it's time to connect to the cluster with `kubectl`. This will require
obtaining the new cluster's certificates adding new `kubectl` configuration.
Here [jq] comes in handy.

```bash
export CLUSTER_NAME="example-cluster"
export COMMON_DOMAIN="internal.company.com"
export CERT_DIR="./certs"

mkdir -p ${CERT_DIR}

kubectl get secret ${CLUSTER_NAME}-api -o json | jq -r .data.ca | base64 --decode > ${CERT_DIR}/ca.crt
kubectl get secret ${CLUSTER_NAME}-api -o json | jq -r .data.crt | base64 --decode > ${CERT_DIR}/apiserver.crt
kubectl get secret ${CLUSTER_NAME}-api -o json | jq -r .data.key | base64 --decode > ${CERT_DIR}/apiserver.key

kubectl config set clusters.${CLUSTER_NAME}.certificate-authority "${CERT_DIR}/ca.crt"
kubectl config set clusters.${CLUSTER_NAME}.server "https://api.${CLUSTER_NAME}.${COMMON_DOMAIN}"
kubectl config set contexts.${CLUSTER_NAME}.cluster "${CLUSTER_NAME}"
kubectl config set contexts.${CLUSTER_NAME}.user "${CLUSTER_NAME}"
kubectl config set users.${CLUSTER_NAME}.client-certificate "${CERT_DIR}/apiserver.crt"
kubectl config set users.${CLUSTER_NAME}.client-key "${CERT_DIR}/apiserver.key"
```

Now with `kubectl` configured let's display `cluster-info`.

```bash
export CLUSTER_NAME="example-cluster"

kubectl config use-context ${CLUSTER_NAME}
kubectl cluster-info
```


## Cleaning Up

First delete the certificate TPOs.

```bash
export CLUSTER_NAME="example-cluster"

kubectl delete aws -l clusterID=${CLUSTER_NAME}
```

Wait for the operator to delete the cluster, and then remove the operator's
deployment and configuration.

```bash
kubectl delete -f ./deployment.yaml
kubectl delete -f ./configmap.yaml
kubectl delete -f ./configmap-ssh.yaml
```

Finally remove `kubectl` cluster configuration.

```bash
export CLUSTER_NAME="example-cluster"

kubectl config unset clusters.${CLUSTER_NAME}
kubectl config unset contexts.${CLUSTER_NAME}
kubectl config unset users.${CLUSTER_NAME}
```

[aws-operator]: https://github.com/giantswarm/aws-operator
[cert-operator]: https://github.com/giantswarm/cert-operator
[cert-operator-local-setup]: https://github.com/giantswarm/cert-operator/tree/master/examples/local

[jq]: https://stedolan.github.io/jq
