package cloudconfig

const Small = `{
  "ignition": {
    "version": "3.0.0",
    "config": {
      "append": [
        {
          "source": "{{ .S3URL }}"
        }
      ]
    }
  },
  "storage": {
    "filesystems": [
      {
        "path": "/var/lib/docker",
        "device": "/dev/xvdc",
        "wipeFilesystem": true,
        "label": "docker",
        "format": "xfs"
      },
      {
        "path": "/var/log",
        "device": "/dev/xvdf",
        "wipeFilesystem": true,
        "label": "log",
        "format": "xfs"
      },
      {
        "path": "/var/lib/etcd",
        "device": "/dev/xvdh",
        "wipeFilesystem": false,
        "label": "etcd",
        "format": "ext4"
      }
    ]
  }
}
`
