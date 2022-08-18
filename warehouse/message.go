package warehouse

import (
	"Memcached/internal"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

const (
	Md5Size = 16
)

type Message struct {
	Key       []byte   `json:"key"`
	Md5       [16]byte `json:"md5"`
	Timestamp int64    `json:"timestamp,omitempty"`
	HoldTime  int16    `json:"hold_time"`
	Body      []byte   `json:"body"`
	kt        *keyItem
}

func NewMessage(key, md5Value []byte, body []byte, holdTime int16) (*Message, error) {
	md5Check := md5.Sum(body)
	if bytes.Equal(md5Check[:], md5Value) {
		return nil, errors.New("md5 not same")
	}
	return &Message{
		Key:       key,
		Md5:       md5Check,
		HoldTime:  holdTime,
		Timestamp: time.Now().UnixNano(),
		Body:      body,
	}, nil
}

func NewMessageByStr(key, md5Value, body string, holdTime int16) (*Message, error) {
	md5Check := md5.Sum([]byte(body))
	if md5Value != fmt.Sprintf("%x", md5Check) {
		return nil, errors.New("md5 not same")
	}
	return &Message{
		Key:       []byte(key),
		Md5:       md5Check,
		HoldTime:  holdTime,
		Timestamp: time.Now().UnixNano(),
		Body:      []byte(body),
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
	if err := binary.Write(b, binary.BigEndian, m.HoldTime); err != nil {
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
	if err := binary.Read(bf, binary.BigEndian, &msg.HoldTime); err != nil {
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
