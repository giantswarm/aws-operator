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
  }
}
`
