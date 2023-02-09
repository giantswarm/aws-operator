package template

const ImagePuller = `#!/bin/bash

token="$(cat /root/.docker/config.json |jq '.auths["https://index.docker.io/v1/"].auth' -r)"
crictl pull --auth $token $1
`
