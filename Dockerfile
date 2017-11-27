FROM busybox:ubuntu-14.04

RUN mkdir -p /opt
ADD ./aws-operator /opt/aws-operator

RUN mkdir -p /service
ADD service/templates/ /service/templates

WORKDIR /opt

EXPOSE 8000
ENTRYPOINT ["/opt/aws-operator"]
