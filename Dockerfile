FROM busybox:ubuntu-14.04

RUN mkdir -p /opt
ADD ./aws-operator /opt/aws-operator

EXPOSE 8000
ENTRYPOINT ["/opt/aws-operator"]
