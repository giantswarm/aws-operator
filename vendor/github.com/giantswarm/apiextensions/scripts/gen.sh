#!/usr/bin/env bash

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# TODO Workaround https://github.com/kubernetes/code-generator/issues/6
# boilerplate.go.txt is copied from
# https://github.com/kubernetes/kubernetes/blob/45db5e7260d47f01106c3f0401f9c779f9b386c0/hack/boilerplate/boilerplate.go.txt
if [[ -z $GOPATH ]]; then
    echo "GOPATH env var must be set" >&2
    exit 1
fi
if [[ ! -f ${GOPATH}/src/k8s.io/kubernetes/hack/boilerplate/boilerplate.go.txt ]]; then
    mkdir -p ${GOPATH}/src/k8s.io/kubernetes/hack/boilerplate
    cp ${dir}/boilerplate.go.txt ${GOPATH}/src/k8s.io/kubernetes/hack/boilerplate/boilerplate.go.txt
fi

cd ${dir}/../vendor/k8s.io/code-generator && ./generate-groups.sh \
    all \
    github.com/giantswarm/apiextensions/pkg \
    github.com/giantswarm/apiextensions/pkg/apis \
    "core:v1alpha1 provider:v1alpha1"

for f in $(find ${dir}/../pkg/* -name "*_client.go"); do 
    sed -i $f -e 's,"/api","/apis",g'
done
