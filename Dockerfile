FROM golang:1.13 AS builder
ENV GO111MODULE=on
COPY go.mod /etc/go.mod
RUN cat /etc/go.mod | grep k8scloudconfig | awk '{print $1"@"$2}' | xargs -I{} go get {}

FROM alpine:3.8

RUN apk add --no-cache ca-certificates

RUN mkdir -p /opt/aws-operator
ADD ./aws-operator /opt/aws-operator/aws-operator

RUN mkdir -p /opt/ignition
COPY --from=builder /go/pkg/mod/cache/download/github.com/giantswarm/k8scloudconfig /opt/ignition

WORKDIR /opt/aws-operator

EXPOSE 8000
ENTRYPOINT ["/opt/aws-operator/aws-operator"]
