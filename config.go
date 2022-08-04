package Memcached

import (
	"time"
)

type Config struct {
	TcpServerAddr string
	RemoteAddrArr []string
	StoreCap      int

	SyncCheck time.Duration
}
