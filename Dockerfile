FROM golang:1.17.4 AS builder
ENV GO111MODULE=on
COPY go.mod /etc/go.mod
RUN cat /etc/go.mod | grep k8scloudconfig | awk '{print $1"/...@"$2}' | xargs -I{} go get {}
# This is needed to extract the versioned catalog name, e.g. v6@6.0.1
RUN ln -s /go/pkg/mod/$(cat /etc/go.mod | grep k8scloudconfig | awk '{print $1"@"$2}') /opt/k8scloudconfig

FROM alpine:3.15.0

RUN apk add --no-cache ca-certificates

RUN mkdir -p /opt/aws-operator
ADD ./aws-operator /opt/aws-operator/aws-operator

RUN mkdir -p /opt/ignition
COPY --from=builder /opt/k8scloudconfig /opt/ignition

WORKDIR /opt/aws-operator

EXPOSE 8000
ENTRYPOINT ["/opt/aws-operator/aws-operator"]
