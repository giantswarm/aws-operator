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
        "name": "docker",
        "mount": {
          "device": "{{if eq .InstanceRole "master"}}/dev/xvdc{{else}}/dev/xvdh{{end}}",
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
      }{{ if eq .InstanceRole "worker" -}},
      {
        "name": "kubelet",
        "mount": {
          "device": "/dev/xvdg",
          "wipeFilesystem": true,
          "label": "kubelet",
          "format": "xfs"
        }
      }
      {{- end }}{{ if eq .InstanceRole "master" -}},
      {
        "name": "etcd",
        "mount": {
          "device": "/dev/xvdh",
          "wipeFilesystem": false,
          "label": "etcd",
          "format": "ext4"
        }
      }
    {{- end }}
    ]
  }
}
`
