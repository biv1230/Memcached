package command

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
)

func ReceiveFile(name, size []byte, r io.Reader) error {
	path := "D:\\store\\recv\\"
	file, err := os.Create(path + string(name))
	defer file.Close()
	if err != nil {
		return err
	}
	fileSize := int(binary.BigEndian.Uint64(size))
	log.Printf("开始接受文件[%s],大小[%d]\n", name, fileSize)
	buf := make([]byte, 4096)
	nr, nw := 0, 0
	for {
		n, err := r.Read(buf)
		if err != nil {
			log.Println("read:", err)
			return err
		}
		nr += n
		buf = buf[:n]
		n2, err := file.Write(buf)
		if err != nil {
			log.Println("write:", err)
		}
		nw += n2
		//log.Println("文件传输进度:", fileSize, nr, nw)
		if nr == fileSize {
			break
		}
		file.Sync()
	}
	if nw == fileSize {
		log.Printf("结束文件[%s],大小[%d],接受[%d],写入[%d]\n", name, fileSize, nr, nw)
		return nil
	} else {
		return fmt.Errorf("文件大小[%d],接受[%d],写入[%d]", fileSize, nr, nw)
	}
}
