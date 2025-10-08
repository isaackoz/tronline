package signaling

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Room struct {
	ID     string
	Host   *Client
	Guest  *Client
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
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

// runs a rooms logic/loop. rooms have a lifetime of 10 mins
// the room is just to connect a host <-> guest. Once they're connected, the room will close
// and the two clients will communicate P2P via WebRTC from thereon
func (r *Room) Run(rootCtx context.Context) error {
	// rooms will close automatically after 10 mins
	r.ctx, r.cancel = context.WithTimeout(rootCtx, time.Minute*10)
	defer r.cancel()
	slog.Debug("room running", "room_id", r.ID)

	<-r.ctx.Done()
	slog.Debug("room context done, cleaning up", "room_id", r.ID)
	if err := r.Cleanup(); err != nil {
		slog.Error("cleanup room", "error", err, "room_id", r.ID)
	}
	return nil
}

func (r *Room) AddClient(client *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if client.IsHost {
		if r.Host != nil {
			return &RoomError{Message: "room already has a host"}
		}
		r.Host = client
		slog.Debug("host joined room", "room_id", r.ID, "client_id", client.ID)
	} else {
		if r.Guest != nil {
			return &RoomError{Message: "room already has a guest"}
		}
		r.Guest = client
		slog.Debug("guest joined room", "room_id", r.ID, "client_id", client.ID)
		if r.Host != nil {
			r.Host.SendMessage(&EventMessage{
				Type: MessageEventTypeGuestJoined,
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
			r.Guest.SendMessage(&EventMessage{
				Type:     MessageEventTypeHostLeft,
				Metadata: nil,
			})
		}

		slog.Debug("host left room", "room_id", r.ID, "client_id", client.ID)
		if r.cancel != nil {
			r.cancel() // cancel the room context to trigger cleanup
		}

	} else {
		slog.Debug("client left room", "room_id", r.ID, "client_id", client.ID)
		r.Guest = nil
		if r.Host != nil {
			r.Host.SendMessage(&EventMessage{
				Type: MessageEventTypeGuestLeft,
			})
		}
	}

}

// any message coming from the client will be routed to the other client in the room
// in other words, we just forward the message to the other client and don't handle it here
// note: the messages must still be under 128kb as we defined in the websocket upgrader
func (r *Room) RouteMessage(msg Message, from *Client) {
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
	slog.Debug("routed message", "from", from.ID, "to", target.ID, "room_id", r.ID, "type", msg.GetType())
}

func (r *Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Host == nil && r.Guest == nil
}

// Closes the room, disconnects any clients left
func (r *Room) Cleanup() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Host != nil {
		r.Host.SendMessage(&EventMessage{
			Type: MessageEventRoomClosed,
		})
		close(r.Host.Send)
		r.Host = nil
	}

	if r.Guest != nil {
		r.Guest.SendMessage(&EventMessage{
			Type: MessageEventRoomClosed,
		})
		close(r.Guest.Send)
		r.Guest = nil
	}

	slog.Debug("room cleaned up", "room_id", r.ID)
	return nil
}
