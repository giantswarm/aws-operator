package server

import (
	"github.com/giantswarm/microkit/command/daemon/flag/server/enable"
	"github.com/giantswarm/microkit/command/daemon/flag/server/listen"
	"github.com/giantswarm/microkit/command/daemon/flag/server/log"
	"github.com/giantswarm/microkit/command/daemon/flag/server/tls"
)

type Server struct {
	Enable enable.Enable
	Listen listen.Listen
	Log    log.Log
	TLS    tls.TLS
}
