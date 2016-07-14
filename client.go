package main

import (
	"github.com/gorilla/websocket"
	"time"
)

type client struct {
	socket *websocket.Conn
	send   chan *message
	room   *room
	userData map[string]interface{}
}

func (c *client) read() {
	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err == nil {
			t := time.Now()
			msg.When = t.Format("2006-01-02 15:04:05")//time.UnixDate)
			msg.Name = c.userData["name"].(string)
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}

func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
