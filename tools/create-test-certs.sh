#!/bin/bash

set -e

CERTS_DIR=certs
KEY_SIZE=512
# currently, this is the manually created Elastic IP of the master
CN=35.158.16.27

mkdir -p $CERTS_DIR $CERTS_DIR/calico $CERTS_DIR/etcd

cd $CERTS_DIR || exit 1

keys=(apiserver worker calico/client etcd/server)

for key in ${keys[*]}; do
  # generate server key, unencrypted
  openssl genrsa -out "${key}-key.pem" $KEY_SIZE
  # generate certificate request
  openssl req -new -key "${key}-key.pem" -out "${key}.csr" -subj "/CN=$CN"

  # generate CA
  openssl genrsa -out "${key}-ca-key.pem" $KEY_SIZE
  openssl req -x509 -new -nodes -key "${key}-ca-key.pem" -days 365 -out "${key}-ca.pem" -subj "/CN=$CN"

  # sign certificate
  openssl x509 -req -in "${key}.csr" -CA "${key}-ca.pem" -CAkey "${key}-ca-key.pem" -CAcreateserial -out "${key}.pem"  -days 365
done
