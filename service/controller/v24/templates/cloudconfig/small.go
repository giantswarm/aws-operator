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
    "filesystems": [
      {
        "mount": {
          "device": "{{if eq .InstanceRole "master"}}/dev/xvdc{{else}}/dev/xvdh{{end}}",
          "wipeFilesystem": true,
          "label": "docker",
          "format": "xfs"
        }
      }
    ]
  }
}
`
