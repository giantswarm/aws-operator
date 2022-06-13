FROM golang:1.18.0 AS builder
ENV GO111MODULE=on
COPY go.mod /etc/go.mod
RUN cat /etc/go.mod | grep k8scloudconfig | awk '{print $1"/...@"$2}' | git clone https://github.com/giantswarm/k8scloudconfig.git && cd k8scloudconfig | git checkout $2 | make build | cp k8scloudconfig /opt/k8scloudconfig

FROM alpine:3.15.4

RUN apk add --no-cache ca-certificates

RUN mkdir -p /opt/aws-operator
ADD ./aws-operator /opt/aws-operator/aws-operator

RUN mkdir -p /opt/ignition
COPY --from=builder /opt/k8scloudconfig /opt/ignition

WORKDIR /opt/aws-operator

EXPOSE 8000
ENTRYPOINT ["/opt/aws-operator/aws-operator"]
