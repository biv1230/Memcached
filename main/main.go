package main

import (
	"Memcached"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	cf := &Memcached.Config{
		TcpServerAddr: "127.0.0.1:3001",
		RemoteAddrArr: []string{"127.0.0.1:3001", "127.0.0.1:3002", "127.0.0.1:3003"},
	}
	Memcached.Start(cf, logrus.New())

	time.Sleep(time.Hour)
}
