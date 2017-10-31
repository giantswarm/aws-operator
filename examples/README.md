# Running aws-operator Locally

**Note:** This should only be used for testing and development. See the
[/kubernetes/][kubernetes-dir] directory and [Secrets][secrets-doc] for
a production ready configuration.

[kubernetes-dir]: https://github.com/giantswarm/aws-operator/tree/master/kubernetes
[secrets-doc]: https://github.com/giantswarm/aws-operator#secret

This guide explains how to get aws-operator running locally - on minikube, for
example. Also how to create a managed Kubernetes cluster with single master and
single worker using the operator.

If not stated otherwise all commands are assumed to be run from `examples/local`
directory.

## Cluster Certificates

The easiest way to create certificates is to use the local [cert-operator]
setup. See [this guide][cert-operator-local-setup] for details. Using the [helm registry plugin]
the cert-operator charts can be installed easily:

```bash
helm registry install quay.io/giantswarm/cert-operator-lab-chart -- \
                   -n cert-operator-lab \
                   --set imageTag=latest \
                   --set clusterName=my-cluster \
                   --set commonDomain=my-common-domain \
                   --wait
helm registry install quay.io/giantswarm/cert-resource-lab-chart -- \
                   -n cert-resource-lab \
                   --set clusterName=my-cluster \
                   --set commonDomain=my-common-domain
```

- Note: `clusterName` and `commonDomain` chart values must match the values used
  during this guide.

[helm registry plugin]: https://github.com/app-registry/appr-helm-plugin

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

# From the root of the project, where the Dockerfile resides
GOOS=linux go build github.com/giantswarm/aws-operator
docker build -t quay.io/giantswarm/aws-operator:local-lab .

# Optional. Restart running operator after image update.
# Does nothing when the operator is not deployed.
#kubectl delete pod -l app=aws-operator-local
```

## Deploying the lab charts

The lab consist of two Helm charts, `aws-operator-lab-chart`, which sets up aws-operator,
and `aws-resource-lab-chart`, which defines the cluster to be created.

With a working Helm installation they can be created from the `examples/local` dir with:

```bash
$ helm install -n aws-operator-lab ./aws-operator-lab-chart/ --wait
$ helm install -n aws-resource-lab ./aws-resource-lab-chart/ --wait
```

`aws-operator-lab-chart` accepts the following configuration parameters:
* `idRsaPub` - SSH public key to be installed on nodes.
* `aws.accessKeyId` - AWS access key.
* `aws.secretAccessKey` - AWS secret.
* `aws.sessionToken` - AWS session token for MFA accounts; can be left empty.
* `imgeTag` - Tag of the aws-operator image to be used, by default `local-lab` to use a
locally created image.

For instance, to pass your default ssh public key to the install command, along with AWS
credentials from the environment, you could do:

```bash
$ helm install -n aws-operator-lab --set idRsaPub="$(cat ~/.ssh/id_rsa.pub)" \
                                   --set aws.accessKeyId=${AWS_ACCESS_KEY_ID} \
                                   --set aws.secretAccessKey=${AWS_SECRET_ACCESS_KEY} \
                                   --set aws.sessionToken=${AWS_SESSION_TOKEN} \
                                   ./aws-operator-lab-chart/ --wait
```

`aws-resource-lab-chart` accepts the following configuration parameters:
* `clusterName` - Cluster's name.
* `commonDomain` - Cluster's etcd and API common domain.
* `sshUser` - SSH user created via cloudconfig.
* `sshPublicKey` - SSH public key added via cloudconfig.
* `aws.region` - AWS region.
* `aws.az` - AWS availability zone.
* `aws.ami` - AWS image to be used on both master and worker machines.
* `aws.instanceTypeMaster` - Master machines instance type.
* `aws.instanceTypeWorker` - Worker machines instance type.
* `aws.apiHostedZone` - Route 53 hosted zone for API and Etcd
* `aws.ingressHostedZone` - Route 53 hosted zone for Ingress
* `aws.routeTable0` - Existing route table of the cluster to use for VPC peering.
* `aws.routeTable1` - Existing route table of the cluster to use for VPC peering.
* `aws.vpcPeerId` - VPC ID of the host cluster to peer with.

For instance, to create a SSH user with your current user and default public key.

```bash
$ helm install -n aws-resource-lab --set sshUser="$(whoami)" \
                                  --set sshPublicKey="$(cat ~/.ssh/id_rsa.pub)" \
                                   ./aws-resource-lab-chart/ --wait
```

## Connecting to the new cluster

To test if the cluster is ready check the operator's pod logs with the
command below.

```bash
kubectl logs -l app=aws-operator-local
```

When a similar message appears in the log output, the cluster is ready.

```
{"caller":"github.com/giantswarm/aws-operator/service/create/service.go:967","info":"cluster 'test-cluster' processed","time":"17-05-24 15:24:08.537"}
```

Now it's time to connect to the cluster with `kubectl`. This will require
obtaining the new cluster's certificates adding new `kubectl` configuration.
Here [jq] comes in handy.

```bash
export CLUSTER_NAME="test-cluster"
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

First delete the cluster TPO.

```bash
$ helm delete aws-resource-lab --purge
```

Wait for the operator to delete the cluster, you should see a message like this in the operator logs:

```
{"caller":"github.com/giantswarm/aws-operator/service/create/service.go:967","info":"cluster 'test-cluster' deleted","time":"17-05-24 15:24:08.537"}
```

Then remove the operator's deployment and configuration.

```bash
$ helm delete aws-operator-lab --purge
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
[cert-operator-local-setup]: https://github.com/giantswarm/cert-operator/tree/master/examples

[jq]: https://stedolan.github.io/jq
