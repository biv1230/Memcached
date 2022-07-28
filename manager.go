package Memcached

import (
	"Memcached/cache"
	"Memcached/internal"
	"Memcached/server"
	"context"
)

type Manager struct {
	c *cache.Caches
	t *server.TcpConnects
}

func Start(cf *Config, lg internal.Logger) {
	internal.Lg = lg
	_, err := server.CMStart(context.Background(), cf.newServerConfig())
	if err != nil {
		lg.Errorf("start error:[%s]", err)
	}

}
