package cache

import (
	"bytes"
	"crypto/md5"
	"testing"
	"time"
)

func TestMessage(t *testing.T) {
	msg := &Message{
		Key:       []byte("中国OK"),
		Body:      []byte("这个是测试,谢谢"),
		Timestamp: time.Now().UnixNano(),
	}
	msg.Md5 = md5.Sum(msg.Body)

	b, err := msg.ToByte()
	if err != nil {
		t.Errorf("message to []byte err: %s", err)
	}
	ret, err := DecodeMessage(b)
	if err != nil {
		t.Errorf("message decode []byte err: %s", err)
	}
	if !bytes.Equal(msg.Key, ret.Key) {
		t.Errorf("key not same: %s, %s", msg.Key, ret.Key)
	}
	if !bytes.Equal(msg.Md5[:], ret.Md5[:]) {
		t.Errorf("md5 not same: %x, %x", msg.Md5, ret.Md5)
	}
	if ret.Timestamp != msg.Timestamp {
		t.Errorf("timestamp not same: %d, %d", msg.Timestamp, ret.Timestamp)
	}
	if !bytes.Equal(msg.Body, ret.Body) {
		t.Errorf("md5 not same: %s, %s", msg.Body, ret.Body)
	}
}
