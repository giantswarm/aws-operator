FROM quay.io/giantswarm/golang:1.14.1 AS builder

WORKDIR /mod

ADD go.mod .

RUN K8SCCPATH=$(go list -m -f '{{.Path}}' github.com/giantswarm/k8scloudconfig/...) \
    && go mod download -x $K8SCCPATH \
    && K8SCCROOT=$(go list -m -f '{{.Dir}}' github.com/giantswarm/k8scloudconfig/...) \
    && mv $K8SCCROOT/files /opt/k8scloudconfig

FROM alpine:3.8

RUN apk add --no-cache ca-certificates

RUN mkdir -p /opt/aws-operator
ADD ./aws-operator /opt/aws-operator/aws-operator

RUN mkdir -p /opt/ignition
COPY --from=builder /opt/k8scloudconfig /opt/ignition

WORKDIR /opt/aws-operator

EXPOSE 8000
ENTRYPOINT ["/opt/aws-operator/aws-operator"]
