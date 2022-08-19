package main

import (
	"Memcached"
	"Memcached/internal"
	"Memcached/warehouse"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Message struct {
	Key      string `json:"key"`
	Body     string `json:"body"`
	HoldTime int16  `json:"hold_time"`
	Md5      string `json:"md5"`
}

func httpStart() {
	h := gin.Default()

	h.POST("/save_data", func(c *gin.Context) {
		body := Message{}
		err := c.BindJSON(&body)
		if err != nil {
			internal.Lg.Errorf("parse body err:[%s]", err)
			c.JSON(http.StatusOK, gin.H{"err": err.Error()})
			return
		}
		nMsg, err := warehouse.NewMessageByStr(body.Key, body.Md5, body.Body, body.HoldTime)
		if err != nil {
			internal.Lg.Errorf("new msg err:[%s]", err)
			c.JSON(http.StatusOK, gin.H{"err": err.Error()})
			return
		}
		err = Memcached.SaveMessage(nMsg)
		c.JSON(http.StatusOK, gin.H{"msg": "suc", "code": "suc"})
	})

	h.GET("/get/:key", func(c *gin.Context) {
		key := c.Param("key")
		msg := Memcached.GetMessage([]byte(key))
		var context string
		if msg == nil {
			context = ""
		} else {
			context = string(msg.Body)
		}
		c.JSON(http.StatusOK, gin.H{"body": context})
	})

	if err := h.Run(":3636"); err != nil {
		internal.Lg.Errorf("gin run err:[%s]", err)
	}
}

func main() {
	cf := &Memcached.Config{
		SyncCheck:     10 * time.Second,
		StoreCap:      10,
		TcpServerAddr: "127.0.0.1:3001",
		RemoteAddrArr: []string{"127.0.0.1:3001", "127.0.0.1:3002", "127.0.0.1:3003"},
	}
	Memcached.Start(context.Background(), cf, Memcached.Log)

	httpStart()
}
