[![CircleCI](https://circleci.com/gh/giantswarm/k8scloudconfig.svg?&style=shield&circle-token=d82e253ec55ee80292084262e2c022c442797fd0)](https://circleci.com/gh/giantswarm/k8scloudconfig)

# k8scloudconfig
Cloud-init configuration for setting up Kubernetes clusters

## Versioning

k8scloudconfig library uses semver versioning scheme. Please follow simple rules, when creating new version:

1. Increment MAJOR version number when breaking library API changes introduced.
2. Increment PATCH version number for critical bug fixes. Patch release needs to be immediately included into patch release of operator.
3. Increment MINOR version number for all other changes.
4. WIP releases are only possible for major and minor version updates. Patch releases should be immediately frozen.

Examples:
- "Hyperkube upgrade from 1.9.5 to 1.10.1" is a minor version upgrade.
- "New field `DisableCalico` added to `Params` struct" is a major version upgrade.
- "Kubelet configuration changed to prevent stuck in terminating state pods" is a patch version upgrade.

## Branches

- `legacy`
    - All versions matching ~0.1.0.
    - For dep-based operators.
- `master`
    - From version v6.0.0.
    - For operators using go modules.
