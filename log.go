package Memcached

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"
)

var (
	formatStr      = "[%s][%s][%s][%s][%s]"
	defaultTimeStr = "20060102 15:04:05"
	Log            = log.New()
)

func init() {
	Log.SetFormatter(&TLFormat{})
	Log.SetReportCaller(true)
}

func SetLogLevel(level uint32) {
	Log.SetLevel(log.Level(level))
}

type TLFormat struct {
}

func (m *TLFormat) Format(e *log.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if e.Buffer != nil {
		b = e.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	funcVal, fileVal := getFuncFile(e)
	b.WriteString(fmt.Sprintf(formatStr, e.Level.String(), e.Time.Format(defaultTimeStr), e.Message, funcVal, fileVal))
	for k, v := range e.Data {
		m.appendKeyValue(b, k, v)
	}
	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (m *TLFormat) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	b.WriteByte(' ')
	b.WriteString(key)
	b.WriteByte('=')
	b.WriteString(fmt.Sprint(value))
}

func getFuncFile(entry *log.Entry) (funcVal string, fileVal string) {
	if entry.HasCaller() {
		funcVal = entry.Caller.Function
		fileVal = fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
	}
	return
}
