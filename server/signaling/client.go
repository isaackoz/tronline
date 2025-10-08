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
	pongWait       = time.Second * 25
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 * 128 // 128KB
)

type Client struct {
	ID     string
	Hub    *Hub
	Room   *Room
	Conn   *websocket.Conn
	Send   chan []byte // for sending messages to the client. buffer of 256 messages
	IsHost bool
	Ctx    context.Context
}

// read and write messages to/from the websocket connection. this is client<->server
func (c *Client) ReadWriteWs(ctx context.Context) {
	c.Conn.SetReadLimit(maxMessageSize)
	readChan := make(chan Message, 10)
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		slog.Debug("Closing client connection in readwritews", "client_id", c.ID)
		c.Conn.Close(websocket.StatusNormalClosure, "closing")
		pingTicker.Stop()
	}()

	go func(conn *websocket.Conn) {
		if conn == nil {
			slog.Error("websocket connection is nil in read goroutine", "client_id", c.ID)
			return
		}
		defer close(readChan)
		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				var wsErr websocket.CloseError
				if errors.As(err, &wsErr) {
					if wsErr.Code != websocket.StatusNormalClosure && wsErr.Code != websocket.StatusGoingAway {
						slog.Error("websocket closed unexpectedly", "code", wsErr.Code, "reason", wsErr.Reason)
					}
				} else if !errors.Is(err, context.Canceled) {
					slog.Error("read message", "error", err)
				}
				// otherwise it was a normal closure
				return
			}
			var msg Message
			if err := json.Unmarshal(data, &msg); err != nil {
				slog.Error("unmarshal message", "error", err)
				continue
			}
			//todo parse/validate the message, convert to the appropriate message type, and then forward to other peer
			// if c.IsHost {
			// 	msg.From = "host"
			// } else {
			// 	msg.From = "client"
			// }
			select {
			case readChan <- msg:
			default:
				slog.Debug("read channel full, dropping message", "client_id", c.ID)
			}
		}
	}(c.Conn)

	for {
		select {
		case msg, ok := <-readChan:
			if !ok {
				slog.Debug("read channel closed, closing connection", "client_id", c.ID)
				return
			}
			slog.Debug("message received", "type", msg.GetType(), "client_id", c.ID)
			msgType := msg.GetType()
			switch msgType {
			case MessageTypeOffer, MessageTypeAnswer, MessageTypeICECandidate:
				c.Room.RouteMessage(msg, c)
			case MessageTypeWebRTCConnected:
				// when the client tells us that they connected to their peer, our work is done here
				slog.Debug("webrtc connected", "client_id", c.ID, "room_id", c.Room.ID)
				// TODO: close the websocket connection as we don't need it anymore
				return
			default:
				slog.Warn("unknown message type", "type", msg.GetType(), "client_id", c.ID)
			}
		case writeMessage, ok := <-c.Send:
			if !ok {
				slog.Debug("send channel closed, closing connection", "client_id", c.ID)
				return
			}
			err := c.Conn.Write(ctx, websocket.MessageText, writeMessage)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					slog.Debug("writing to client", "error", err)
				}
				return
			}
		case <-pingTicker.C:
			if err := c.Conn.Ping(ctx); err != nil {
				if !errors.Is(err, context.Canceled) {
					slog.Debug("pinging client", "error", err)
				}
				return
			}
		case <-ctx.Done():
			slog.Debug("client context done, closing connection", "client_id", c.ID)
			return
		}
	}
}

func (c *Client) SendMessage(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("marshal message", "error", err)
		return
	}

	select {
	case c.Send <- data:
		slog.Debug("sent message", "type", msg.GetType(), "client_id", c.ID)
	default:
		slog.Debug("send channel full, closing connection", "client_id", c.ID)
		close(c.Send)
	}
}
