package cloudconfig

const InstanceStorage = `
storage:
  filesystems:
    - name: ephemeral1
      mount:
        device: /dev/xvdb
        format: xfs
        create:
          force: true
`
