package command

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

func sendFileInfoStatus(name string, size int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(size))
	params := [][]byte{FileInfoCommandByte, []byte(name), b}
	return bytes.Join(params, ByteSpace)
}

func SendFile(path string, w io.Writer) error {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	statsComm := sendFileInfoStatus(fileInfo.Name(), fileInfo.Size())
	if _, err := w.Write(statsComm); err != nil {
		return err
	}
	if _, err := w.Write([]byte{NewLine}); err != nil {
		return err
	}
	if n, err := bufio.NewWriter(w).ReadFrom(file); err != nil || n != fileInfo.Size() {
		//发送一个文件传输有的消息？
		return err
	}
	return nil
}
