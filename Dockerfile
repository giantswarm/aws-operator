FROM golang:1.19.0 AS builder
ENV GO111MODULE=on
COPY go.mod /etc/go.mod
RUN git clone --depth 1 --branch $(cat /etc/go.mod | grep k8scloudconfig | awk '{print $2}') https://github.com/giantswarm/k8scloudconfig.git && cp -r k8scloudconfig /opt/k8scloudconfig

FROM alpine:3.16.2

RUN apk add --no-cache ca-certificates

RUN mkdir -p /opt/aws-operator
ADD ./aws-operator /opt/aws-operator/aws-operator

RUN mkdir -p /opt/ignition
COPY --from=builder /opt/k8scloudconfig /opt/ignition

WORKDIR /opt/aws-operator

EXPOSE 8000
ENTRYPOINT ["/opt/aws-operator/aws-operator"]
