# Running aws-operator Locally

**Note:** This should only be used for testing and development. See the
[/kubernetes/][kubernetes-dir] directory and [Secrets][secrets-doc] for
a production ready configuration.

[kubernetes-dir]: https://github.com/giantswarm/aws-operator/tree/master/kubernetes
[secrets-doc]: https://github.com/giantswarm/aws-operator#secret

This guide explains how to get running aws-operator locally. For example on
minikube. Also how to create a managed Kubernetes cluster with single master
and single worker using the operator.

All commands are assumed to be run from `examples/local` directory.


## Preparing Templates

All yaml files in this directory are templates. Before proceeding this guide
all placeholders must be replaced with sensible values.

- *CLUSTER_NAME* - Cluster's name.
- *COMMON_DOMAIN* - Cluster's etcd and API common domain.
- *COMMON_DOMAIN_INGRESS* - Ingress common domain.
- *ID_RSA_PUB* - SSH public key to be installed on nodes.
- *AWS_ACCESS_KEY_ID* - AWS access key.
- *AWS_SECRET_ACCESS_KEY* - AWS secret.
- *AWS_REGION* - AWS region.
- *AWS_AZ* - AWS availability zone.
- *AWS_AMI* - AWS image to be used on both master and worker machines.
- *AWS_INSTANCE_TYPE_MASTER* - Master machines instance type.
- *AWS_INSTANCE_TYPE_WORKER* - Worker machines instance type.

Below is handy snippet than can be used to make that painless. It works in bash and zsh.

```bash
for f in *.tmpl.yaml; do
    sed \
        -e 's/${CLUSTER_NAME}/example-cluster/g' \
        -e 's/${COMMON_DOMAIN}/internal.company.com/g' \
        -e 's/${COMMON_DOMAIN_INGRESS}/company.com/g' \
        -e 's/${ID_RSA_PUB}/ssh-rsa AAsd user@company.com/g' \
        -e 's/${AWS_ACCESS_KEY_ID}/AKIAIXXXXXXXXXXXXXXX/g' \
        -e 's/${AWS_SECRET_ACCESS_KEY}/XXXXXXXXXXXXXXXXX\/XXXXXXXXXXXXXXXXXXXXXX/g' \
        -e 's/${AWS_REGION}/eu-central-1/g' \
        -e 's/${AWS_AZ}/eu-central-1a/g' \
        -e 's/${AWS_AMI}/ami-d60ad6b9/g' \
        -e 's/${AWS_INSTANCE_TYPE_MASTER}/t2.medium/g' \
        -e 's/${AWS_INSTANCE_TYPE_WORKER}/t2.medium/g' \
        ./$f > ./${f%.tmpl.yaml}.yaml
done
```

- Note: Single quotes are intentional. Strings like `${CLUSTER_NAME}` shouldn't
  be interpolated. These are placeholders in the template files.
- Note: `/` inside `AWS_SECRET_ACCESS_KEY` must be escaped.


## Cluster Certificates

The easiest way to create certificates is to use local [cert-operator] setup.
See [this guide][cert-operator-local-setup] for details.

- Note: `CLUSTER_NAME` and `COMMON_DOMAIN` values must match ones used during
  this guide.

## Cluster-Local Docker Image

The operator needs a connection to the K8s API. The simplest approach is to run
as a deployment and use the "in cluster" configuration.

In that case the Docker image needs to be accessible from the K8s cluster
running the operator. For Minikube `eval $(minikube docker-env)` before `docker
build`, see [reusing the Docker daemon] for details.

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

One way is to provide these with ConfigMaps. Please read introduction of this
guide if you want to do it more securely.

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

Creating ThirdPartyObject should eventually result in working K8s cluster on
AWS. To learn if the cluster is ready check the operator's pod logs with the
command below.

```bash
kubectl logs -l app=aws-operator-local
```

When log like this appears in the output the cluster is ready.

```
{"caller":"github.com/giantswarm/aws-operator/service/create/service.go:967","info":"cluster 'pawel' processed","time":"17-05-24 15:24:08.537"}
```

Now it's time to connect the cluster with `kubectl`. This will require
obtaining the new cluster's certificates. If [cert-operator] was used to create
certificates [jq] comes in handy.

```bash
export CLUSTER_NAME="example-cluster"
export CERT_DIR="/cert/path"

kubectl get secret ${CLUSTER_NAME}-api -o json | jq -r .data.ca | base64 --decode > ${CERT_DIR}/ca.crt
kubectl get secret ${CLUSTER_NAME}-api -o json | jq -r .data.crt | base64 --decode > ${CERT_DIR}/apiserver.crt
kubectl get secret ${CLUSTER_NAME}-api -o json | jq -r .data.key | base64 --decode > ${CERT_DIR}/apiserver.key
```

Now let's add following configuration to `~/.kube/config`. Remember to change
`${CLUSTER_NAME}`, `${COMMON_DOMAIN}`, and `${CERT_DIR}`.

```
clusters:
- name: ${CLUSTER_NAME}
  cluster:
    certificate-authority: ${CERT_DIR}/ca.crt
    server: https://api.${CLUSTER_NAME}.${COMMON_DOMAIN}:443

contexts:
- name: ${CLUSTER_NAME}
  context:
    cluster: ${CLUSTER_NAME}
    user: ${CLUSTER_NAME}

users:
- name: ${CLUSTER_NAME}
  user:
    client-certificate: ${CERT_DIR}/apiserver.crt
    client-key: ${CERT_DIR}/apiserver.key
```

It's time to connect to the newly created cluster.

```bash
export CLUSTER_NAME="example-cluster"

kubectl config use-context ${CLUSTER_NAME}
kubectl cluster-info
```


## Cleaning Up

Delete the certificate TPOs and the deployment.

```bash
kubectl delete aws -l clusterID=${CLUSTER_NAME}
```

```
kubectl delete -f ./deployment.yaml
```

[aws-operator]: https://github.com/giantswarm/aws-operator
[cert-operator]: https://github.com/giantswarm/cert-operator
[cert-operator-local-setup]: https://github.com/giantswarm/cert-operator/tree/master/examples/local

[jq]: https://stedolan.github.io/jq
