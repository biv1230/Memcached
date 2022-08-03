package Memcached

import (
	"Memcached/internal"
	"Memcached/server"
	"context"
)

type Manager struct {
	t *server.TcpConnects
}

func Start(cf *Config, lg internal.Logger) {
	internal.Lg = lg
	_, err := server.CMStart(context.Background(), cf.newServerConfig())
	if err != nil {
		lg.Errorf("start error:[%s]", err)
	}
}
