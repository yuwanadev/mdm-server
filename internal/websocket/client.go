package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	ws "github.com/gofiber/contrib/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 4 * 1024 * 1024 // 4MB
)

// Client wraps a WebSocket connection for a single device.
type Client struct {
	DeviceID   uuid.UUID
	DeviceName string
	conn       *ws.Conn
	hub        *Hub
	send       chan []byte
	done       chan struct{}
	closeOnce  sync.Once
}

// NewClient creates a new device client.
func NewClient(conn *ws.Conn, deviceID uuid.UUID, deviceName string, hub *Hub) *Client {
	return &Client{
		DeviceID:   deviceID,
		DeviceName: deviceName,
		conn:       conn,
		hub:        hub,
		send:       make(chan []byte, 64),
		done:       make(chan struct{}),
	}
}

// Start begins the read and write pumps.
func (c *Client) Start() {
	go c.writePump()
	c.readPump() // blocks until disconnect
}

// Send queues a message for delivery to the device.
func (c *Client) Send(msg *WSMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
		return nil
	case <-c.done:
		return ErrDeviceNotConnected
	default:
		log.Printf("[Client:%s] Send buffer full — dropping message", c.DeviceID)
		return nil
	}
}

// Close terminates the client connection.
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.done)
		c.conn.Close()
	})
}

// readPump reads messages from the WebSocket connection.
func (c *Client) readPump() {
	defer func() {
		c.hub.UnregisterDevice(c.DeviceID)
		c.Close()
	}()

	c.conn.SetReadLimit(maxMsgSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseNormalClosure) {
				log.Printf("[Client:%s] Read error: %v", c.DeviceID, err)
			}
			return
		}

		var msg WSMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("[Client:%s] Invalid message: %v", c.DeviceID, err)
			continue
		}

		c.hub.HandleDeviceMessage(c.DeviceID, &msg)
	}
}

// writePump writes messages from the send channel to the WebSocket.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case data, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(ws.TextMessage, data); err != nil {
				log.Printf("[Client:%s] Write error: %v", c.DeviceID, err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(ws.PingMessage, nil); err != nil {
				return
			}

		case <-c.done:
			return
		}
	}
}
