package fileTransfer

import (
	"log"
	"net"
)

func StartServer(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panicf("listen [%s] err: [%s]", addr, err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Panicf("tcp accept err: [%s]", err)
		}
		cl := NewConnChannel(conn)
		go cl.IoLoop()
	}
}
