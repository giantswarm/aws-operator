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
        "name": "docker",
        "mount": {
          "device": "/dev/xvdc",
          "wipeFilesystem": true,
          "label": "docker",
          "format": "xfs"
        }
      },
      {
        "name": "log",
        "mount": {
          "device": "/dev/xvdf",
          "wipeFilesystem": true,
          "label": "log",
          "format": "xfs"
        }
      },
      {
        "name": "etcd",
        "mount": {
          "device": "/dev/xvdh",
          "wipeFilesystem": false,
          "label": "etcd",
          "format": "ext4"
        }
      }
    ]
  }
}
`
