package Memcached

import (
	"Memcached/internal"
	"Memcached/server"
)

type CacheManager struct {
	Log       internal.Logger
	TcpServer *server.TcpServer
}
