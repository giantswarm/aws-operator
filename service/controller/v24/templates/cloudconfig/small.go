package cloudconfig

const Small = `{
  "ignition": {
    "version": "2.2.0",
    "config": {
      "append": [
        {
          "source": "{{ .S3URL }}"
        }
      ]
    }
  },
  "storage": {
    "disks": [
      {
        "device": "/dev/nvme1n1",
        "partitions": [
          {
            "label": "docker"
          }
        ]
      }
    ],
    "filesystems": [
      {
        "mount": {
          "device": "/dev/disk/by-label/docker",
          "wipeFilesystem": true,
          "label": "docker",
          "format": "xfs"
        }
      }
    ]
  }
}
`
