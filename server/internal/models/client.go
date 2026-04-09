package models

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID        string
	Name      string
	Conn      *websocket.Conn
	Send      chan []byte
	RoomID    string
	UserAgent string

	mu        sync.Mutex
	closeOnce sync.Once
}

const (
	WriteWait      = 10 * time.Second
	PongWait       = 60 * time.Second
	PingPeriod     = (PongWait * 9) / 10
	MaxMessageSize = 64 * 1024
)

func (c *Client) SafeWriteJSON(payload []byte) bool {
	select {
	case c.Send <- payload:
		return true
	default:
		return false
	}
}

func (c *Client) Close() error {
	var err error

	c.closeOnce.Do(func() {
		close(c.Send)

		c.mu.Lock()
		defer c.mu.Unlock()
		err = c.Conn.Close()
	})

	return err
}
