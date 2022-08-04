package server

import (
	"Memcached/internal"
	"bufio"
	"bytes"
	"testing"
)

func TestDecodeMessage(t *testing.T) {
	com := &Command{
		Name:   []byte("notice"),
		Params: [][]byte{[]byte("x"), []byte("y"), []byte("z")},
		Body:   []byte("this is  test"),
	}
	bf := internal.BufferPoolGet()
	defer internal.BufferPoolSet(bf)

	w, f := bufio.NewWriter(bf), bufio.NewReader(bf)

	if n, err := com.WriteTo(w); err != nil {
		t.Errorf("command to []byte err: %d, %s", n, err)
	}
	t.Logf("%d", bf.Len())
	line, err := bf.ReadBytes('\n')
	if err != nil {
		t.Errorf("buffer read err: %s", err)
	}
	newCom, err := decodeBody(f, line[:len(line)-1])
	if err != nil {
		t.Errorf("command decode []byte err: %s", err)
	}
	if !bytes.Equal(com.Name, newCom.Name) {
		t.Errorf("Name not same, %s, %s", com.Name, newCom.Name)
	}

	if comParams, newComParams := bytes.Join(com.Params, []byte{' '}), bytes.Join(newCom.Params, []byte{' '}); !bytes.Equal(comParams, newComParams) {
		t.Log(com.Params, "\n")

		t.Log(newCom.Params, "\n")
		t.Logf("%d, %s\n", len(comParams), comParams)
		t.Logf("%d, %s\n", len(newComParams), newComParams)
		t.Errorf("Params not same, %s %s", comParams, newComParams)
	}

	if !bytes.Equal(com.Body, newCom.Body) {
		t.Errorf("Body not same, %s, %s", com.Body, newCom.Body)
	}

}
