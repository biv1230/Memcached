package Memcached

import (
	"Memcached/internal"
	"Memcached/server"
	"Memcached/warehouse"
	"context"
)

func Start(ctx context.Context, cf *Config, lg internal.Logger) {
	internal.Lg = lg
	server.CMStart(ctx, cf.TcpServerAddr, cf.RemoteAddrArr, cf.SyncCheck)
	warehouse.Start(ctx, cf.StoreCap)
}

func GetMessage(key []byte) *warehouse.Message {
	return warehouse.Cache.Get(key)
}

func SaveMessage(msg *warehouse.Message) error {
	server.Notice(msg)
	warehouse.Cache.Add(msg)
	return nil
}
