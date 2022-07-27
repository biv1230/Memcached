package Memcached

import (
	"Memcached/server"
	"time"
)

type Config struct {
	TcpServerAddr string
	RemoteAddrArr []string

	SyncCheck time.Duration
}

func (cf *Config) newServerConfig() *server.Config {
	return &server.Config{
		TcpServerAddr: cf.TcpServerAddr,
		RemoteAddrArr: cf.RemoteAddrArr,
		SyncCheck:     cf.SyncCheck,
	}
}
