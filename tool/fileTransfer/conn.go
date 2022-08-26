package fileTransfer

import (
	"bufio"
	"fileTransfer/command"
	"log"
	"net"
)

type ConnChannel struct {
	w    *bufio.Writer
	r    *bufio.Reader
	conn net.Conn
}

func NewConnChannel(conn net.Conn) *ConnChannel {
	return &ConnChannel{
		w:    bufio.NewWriter(conn),
		r:    bufio.NewReader(conn),
		conn: conn,
	}
}

func (c *ConnChannel) IoLoop() {
	if err := command.SendFile("D:\\store\\send\\Win_Ser_08_R2_SP1_33in1.iso", c.conn); err != nil {
		log.Println(err)
	}
}
