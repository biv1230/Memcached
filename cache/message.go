package cache

import (
	"Memcached/internal"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"io"
	"time"
)

const (
	Md5Size = 16
)

type Message struct {
	Key            []byte
	Md5            [16]byte
	Timestamp      int64
	ExpirationTime int64
	Body           []byte
	kt             *keyItem
}

func NewMessage(key, md5Value []byte, body []byte, ex time.Duration) (*Message, error) {
	md5Check := md5.Sum(body)
	if bytes.Equal(md5Check[:], md5Value) {
		return nil, errors.New("md5 not same")
	}
	now := time.Now()
	return &Message{
		Key:            key,
		Md5:            md5Check,
		ExpirationTime: now.Add(ex).UnixNano(),
		Timestamp:      now.UnixNano(),
		Body:           body,
	}, nil
}

func (m *Message) ToByte() ([]byte, error) {
	b := internal.BufferPoolGet()
	defer internal.BufferPoolSet(b)
	if err := binary.Write(b, binary.BigEndian, uint8(len(m.Key))); err != nil {
		return nil, err
	}
	if _, err := b.Write(m.Key); err != nil {
		return nil, err
	}
	if _, err := b.Write(m.Md5[:]); err != nil {
		return nil, err
	}
	if err := binary.Write(b, binary.BigEndian, m.Timestamp); err != nil {
		return nil, err
	}
	if err := binary.Write(b, binary.BigEndian, uint32(len(m.Body))); err != nil {
		return nil, err
	}
	if _, err := b.Write(m.Body); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func DecodeMessage(b []byte) (*Message, error) {
	bf := internal.BufferPoolGet()
	defer internal.BufferPoolSet(bf)
	bf.Write(b)
	var l uint8
	if err := binary.Read(bf, binary.BigEndian, &l); err != nil {
		return nil, err
	}
	msg := &Message{
		Key: make([]byte, l),
		Md5: [Md5Size]byte{},
	}
	if _, err := io.ReadFull(bf, msg.Key); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(bf, msg.Md5[:]); err != nil {
		return nil, err
	}
	if err := binary.Read(bf, binary.BigEndian, &msg.Timestamp); err != nil {
		return nil, err
	}
	var bodyLen uint32
	if err := binary.Read(bf, binary.BigEndian, &bodyLen); err != nil {
		return nil, err
	}
	if bodyLen > 0 {
		msg.Body = make([]byte, bodyLen)
		if _, err := io.ReadFull(bf, msg.Body); err != nil {
			return nil, err
		}
	}
	return msg, nil
}
