package main

import (
	"Memcached"
	"context"
	"time"
)

func main() {
	cf := &Memcached.Config{
		TcpServerAddr: "127.0.0.1:3001",
		RemoteAddrArr: []string{"127.0.0.1:3001", "127.0.0.1:3002"},
	}
	Memcached.Start(context.Background(), cf, Memcached.Log)

	time.Sleep(time.Hour)
}
