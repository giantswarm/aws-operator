package cloudconfig

const InstanceStorage = `
storage:
  filesystems:
    - name: ephemeral1
      mount:
        device: /dev/xvdb
        format: ext4
        create:
          force: true
`
