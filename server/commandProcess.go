package server

import (
	"fmt"

	"Memcached/internal"
	w "Memcached/warehouse"
)

func ReceiveCommandExec(c *Command) error {
	switch string(c.Name) {
	case CacheBeforeString:
		return ReceiveBefore(c)
	case CachingString:
		return ReceiveSave(c)
	case CacheAfterString:
		return ReceiveAfter(c)
	case PingString:
		return nil
	default:
		return fmt.Errorf("unknow command %s", c.Name)
	}
}

func ReceiveBefore(c *Command) error {
	w.Cache.BeforeAdd(c.Params[0])
	return nil
}

func ReceiveSave(c *Command) error {
	msg, err := w.DecodeMessage(c.Body)
	if err != nil {
		internal.Lg.Errorf("message decode err:[%s]", err)
		return err
	}
	w.Cache.Add(msg)
	return nil
}

func ReceiveAfter(c *Command) error {

	return nil
}
