FROM busybox:ubuntu-14.04

RUN mkdir -p /opt
ADD ./aws-operator /opt/aws-operator

RUN mkdir -p /opt/templates
ADD resources/aws/templates/*.tmpl /opt/templates/

EXPOSE 8000
ENTRYPOINT ["/opt/aws-operator"]
