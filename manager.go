package Memcached

import (
	"Memcached/internal"
	"Memcached/server"
	"Memcached/warehouse"
	"context"
	"time"
)

func Start(ctx context.Context, cf *Config, lg internal.Logger) {
	internal.Lg = lg
	server.CMStart(ctx, cf.TcpServerAddr, cf.RemoteAddrArr, cf.SyncCheck)
	warehouse.Start(ctx, cf.StoreCap)
}

func GetMessage(key []byte) *warehouse.Message {
	return warehouse.Cache.Get(key)
}

func SaveMessage(key, md5 []byte, body []byte, expire time.Duration) error {
	msg, err := warehouse.NewMessage(key, md5, body, expire)
	if err != nil {
		return err
	}
	server.Connects.Notice(msg)
	warehouse.Cache.Add(msg)
	return nil
}
