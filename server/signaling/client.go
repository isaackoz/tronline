package signaling

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/coder/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = time.Second * 60
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 * 128 // 128KB
)

type Client struct {
	ID     string
	Hub    *Hub
	Room   *Room
	Conn   *websocket.Conn
	Send   chan []byte
	IsHost bool
}

// read messages from the websocket connection. client->this server
// todo move this to just go isnide the server
func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.Hub.Unregister <- c
		// ensure we clean up
		c.Conn.CloseNow()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	for {
		slog.Debug("waiting for message", "client_id", c.ID)
		_, message, err := c.Conn.Read(ctx)
		if err != nil {
			var wsErr websocket.CloseError
			if errors.As(err, &wsErr) {
				if wsErr.Code != websocket.StatusNormalClosure {
					slog.Error("websocket closed unexpectedly", "code", wsErr.Code, "reason", wsErr.Reason)
				}
				// otherwise it was a normal closure
			} else {
				// if it was some other error, log it
				slog.Error("read message", "error", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			slog.Error("unmarshal message", "error", err)
			continue
		}

		if c.IsHost {
			msg.From = "host"
		} else {
			msg.From = "client"
		}

		c.Room.RouteMessage(&msg, c)
	}
}

// write messages to the websocket connection. this server->client
func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.CloseNow()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.Close(websocket.StatusNormalClosure, "closed")
				return
			}
			err := c.Conn.Write(ctx, websocket.MessageText, message)
			if err != nil {
				slog.Debug("writing to client", "error", err)
				return
			}
		case <-ticker.C:
			if err := c.Conn.Ping(ctx); err != nil {
				slog.Debug("pinging client", "error", err)
			}
		}
	}
}

func (c *Client) SendMessage(msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("marshal message", "error", err)
		return
	}

	select {
	case c.Send <- data:
	default:
		slog.Debug("send channel full, closing connection", "client_id", c.ID)
		close(c.Send)
	}
}
