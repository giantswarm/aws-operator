FROM alpine:3.7

RUN apk add --no-cache ca-certificates

RUN mkdir -p /opt/aws-operator
ADD ./aws-operator /opt/aws-operator/aws-operator

WORKDIR /opt/aws-operator

EXPOSE 8000
ENTRYPOINT ["/opt/aws-operator/aws-operator"]
