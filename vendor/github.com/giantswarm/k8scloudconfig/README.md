# k8scloudconfig
Cloud-init configuration for setting up Kubernetes clusters

### Keeping bindata in sync

To make sure bindata is always generated and committed when one of the templates
changes, run:

    ln -s ../../tools/update-bindata-hook .git/hooks/pre-commit
