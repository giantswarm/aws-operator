package framework

import (
	"time"
)

const (
	LongMaxWait  = 40 * time.Minute
	ShortMaxWait = 2 * time.Minute
)

const (
	LongMaxInterval  = 60 * time.Second
	ShortMaxInterval = 5 * time.Second
)
