package signaling

import (
	"log/slog"
	"sync"
)

type Room struct {
	ID    string
	Host  *Client
	Guest *Client
	mu    sync.RWMutex
}

func NewRoom(id string) *Room {
	return &Room{
		ID: id,
	}
}

type RoomError struct {
	Message string
}

func (e *RoomError) Error() string {
	return e.Message
}

func (r *Room) AddClient(client *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if client.IsHost {
		if r.Host != nil {
			return &RoomError{Message: "room already has a host"}
		}
		r.Host = client
		r.Host.SendMessage(&Message{
			Type: "room-meta",
			Data: map[string]string{"roomId": r.ID},
		})
		slog.Debug("host joined room", "room_id", r.ID, "client_id", client.ID)
	} else {
		if r.Guest != nil {
			return &RoomError{Message: "room already has a guest"}
		}
		r.Guest = client
		slog.Debug("guest joined room", "room_id", r.ID, "client_id", client.ID)
		if r.Host != nil {
			r.Host.SendMessage(&Message{
				Type: "peer-joined",
			})
		}
	}

	client.Room = r
	return nil
}

func (r *Room) RemoveClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if client.IsHost {
		r.Host = nil

		if r.Guest != nil {
			r.Guest.SendMessage(&Message{
				Type: "peer-left",
			})
		}
	} else {
		r.Guest = nil
		if r.Host != nil {
			r.Host.SendMessage(&Message{
				Type: "peer-left",
			})
		}
	}

	slog.Debug("client left room", "room_id", r.ID, "client_id", client.ID)
}

func (r *Room) RouteMessage(msg *Message, from *Client) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var target *Client
	if from.IsHost {
		target = r.Guest
	} else {
		target = r.Host
	}

	if target == nil {
		slog.Error("no target client to route message to", "from", from.ID, "room_id", r.ID)
		return
	}

	target.SendMessage(msg)
	slog.Debug("routed message", "from", from.ID, "to", target.ID, "room_id", r.ID, "type", msg.Type)

}

func (r *Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Host == nil && r.Guest == nil
}
