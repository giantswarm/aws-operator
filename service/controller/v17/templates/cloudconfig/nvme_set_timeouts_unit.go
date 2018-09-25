package cloudconfig

const NVMESetTimeoutsUnit = `[Unit]
Description=Set NVME timeouts
[Service]
Type=oneshot
ExecStart=/bin/sh -c "\
  [ -d /sys/module/nvme_core/parameters ] && \
  echo 10 > /sys/module/nvme_core/parameters/max_retries && \
  echo 255 > /sys/module/nvme_core/parameters/io_timeout"
[Install]
WantedBy=multi-user.target
`
