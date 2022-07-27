package server

import (
	"Memcached/internal"
	"context"
	"github.com/sirupsen/logrus"
	"testing"
)

var ts *TcpServer

const (
	tcpServer = "127.0.0.1:3001"
)

func init() {
	internal.Lg = logrus.New()
	//ts = NewTcpServer(context.Background(), tcpServer)
	//go ts.Start()
}

func TestConnOtherServer(t *testing.T) {
	_, err := ConnOtherServer(context.Background(), tcpServer, "test1")
	if err != nil {
		t.Errorf("conn err:[%s]", err)
	}
}
