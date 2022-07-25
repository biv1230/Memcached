package main

import (
	"Memcached/server"
	"context"

	"github.com/sirupsen/logrus"
)

func main() {
	ts := server.NewTcpServer(context.Background(), ":0321", logrus.New())
	ts.Start()
}
