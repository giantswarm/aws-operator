FROM busybox:ubuntu-14.04

RUN mkdir -p /opt
ADD ./aws-operator /opt/aws-operator

RUN mkdir -p /opt/resources
ADD resources/templates/ /opt/resources/templates

WORKDIR /opt

EXPOSE 8000
ENTRYPOINT ["/opt/aws-operator"]
