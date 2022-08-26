package fileTransfer

import (
	"bufio"
	"bytes"
	"fileTransfer/command"
	"io"
	"log"
	"net"
)

func ConnectServer() {
	conn, err := net.Dial("tcp", ":6161")
	if err != nil {
		log.Panicf("dial tcp fai [%s]", err)
	}
	defer conn.Close()
	r := bufio.NewReader(conn)
	for {
		line, err := r.ReadSlice(command.NewLine)
		if err != nil {
			if err != io.EOF {
				log.Println(err)
				break
			} else {
				continue
			}
		}
		params := bytes.Split(line[:len(line)-1], command.ByteSpace)
		if bytes.Equal(params[0], command.FileInfoCommandByte) && len(params) == 3 {
			if err := command.ReceiveFile(params[1], params[2], r); err != nil {
				log.Println(err)
			}
		}
	}
}
