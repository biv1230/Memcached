package Memcached

import (
	"Memcached/server"
	"time"
)

type Config struct {
	TcpServerAddr string
	RemoteAddrArr []string

	SyncTimeout time.Duration
}

func (cf *Config) newServerConfig() *server.Config {
	return &server.Config{
		TcpServerAddr: cf.TcpServerAddr,
		RemoteAddrArr: cf.RemoteAddrArr,
		SyncTimeout:   cf.SyncTimeout,
	}
}
