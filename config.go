package Memcached

import (
	"time"
)

type Config struct {
	TcpServerAddr string
	RemoteAddrArr []string
	StoreCap      uint8

	SyncCheck time.Duration
}
