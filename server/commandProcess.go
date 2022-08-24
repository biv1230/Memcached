package server

import (
	"fmt"

	"Memcached/internal"
	w "Memcached/warehouse"
)

func ReceiveCommandExec(client *clientV1, c *Command) error {
	switch string(c.Name) {
	case CacheBeforeString:
		return ReceiveBefore(c)
	case CachingString:
		return ReceiveSave(c)
	case CacheAfterString:
		return ReceiveAfter(c)
	case PingString:
		return nil
	case WarehouseInitString:
		return ReceiveWarehouseInit(client)
	case SendingMsgString:
		return ReceiveSendingMsg(c)
	case SendSucString:
		return ReceiveSendSuc(c)
	default:
		return fmt.Errorf("unknow command %s", c.Name)
	}
}

func ReceiveBefore(c *Command) error {
	w.Cache.BeforeAdd(c.Params[0])
	return nil
}

func ReceiveSave(c *Command) error {
	return w.Cache.AddBody(c.Body)
}

func ReceiveAfter(c *Command) error {
	return nil
}

func ReceiveWarehouseInit(client *clientV1) error {
	l, err := w.Cache.Cap()
	if err != nil {
		return err
	}
	sendCommand := WarehouseInfoCommand(l)
	_, err = sendCommand.WriteTo(client.w)
	if err != nil {
		return err
	}
	err = w.Cache.Range(func(msg *w.Message) error {
		sc, err := SendingMsgCommand(msg)
		if err != nil {
			internal.Lg.Error(err)
			return nil
		} else {
			_, err = sc.WriteTo(client.w)
			return err
		}
	})
	if err != nil {
		_, err = SendFaiCommand.WriteTo(client.w)
	} else {
		_, err = SendSucCommand.WriteTo(client.w)
	}
	return err
}

func ReceiveSendingMsg(c *Command) error {
	return w.Cache.AddBody(c.Body)
}

func ReceiveSendSuc(c *Command) error {
	w.Cache.SetStatus(w.NormalStatus)
	return nil
}
