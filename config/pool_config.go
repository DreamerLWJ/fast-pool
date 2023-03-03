package config

import (
	"time"
)

type PoolConfig struct {
	CorePoolSize    int64
	MaximumPoolSize int64
	KeepAliveTime   time.Duration
	QueueSize       int
}
