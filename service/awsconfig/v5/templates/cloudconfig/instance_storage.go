package cloudconfig

const InstanceStorageTemplate = `
storage:
  filesystems:
    - name: ephemeral1
      mount:
        device: /dev/xvdb
        format: xfs
        create:
          force: true
`
