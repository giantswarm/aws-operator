package cloudconfig

const Small = `{
  "ignition": {
    "version": "2.2.0",
    "config": {
      "append": {
        "source": "{{ .S3URL }}"
      }
    }
  },
  "storage": {
    "filesystems": [
      {
        "mount": {
          "device": "/dev/nvme1n1",
          "format": "xfs",
          "label": "docker"
        },
        "name": "docker"
      },
      {
        "mount": {
          "device": "/dev/nvme2n1",
          "format": "ext4",
          "label": "etcd"
        },
        "name": "var-lib-etcd"
      }
    ]
  }
}
`
