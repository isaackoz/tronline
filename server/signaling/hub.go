package signaling

import (
	"context"
	"log/slog"
	"sync"
)

type Hub struct {
	Rooms map[string]*Room

	Register chan *Client

	Unregister chan *Client

	mu sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)
		case client := <-h.Unregister:
			h.unregisterClient(client)
		case <-ctx.Done():
			slog.Info("hub shutting down")
			return
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.Rooms[client.Room.ID]
	if !exists {
		slog.Debug("room does not exist", "room_id", client.Room.ID)
		client.SendMessage(&Message{
			Type: "error",
			Data: map[string]string{"message": "room not found"},
		})
		close(client.Send)
		return
	}

	if err := room.AddClient(client); err != nil {
		slog.Error("add client to room", "error", err)
		client.SendMessage(&Message{
			Type: "error",
			Data: map[string]string{"message": err.Error()},
		})
		close(client.Send)
		return
	}

	slog.Debug("client registered", "client_id", client.ID, "room_id", room.ID)
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if client.Room != nil {
		client.Room.RemoveClient(client)
	}

	if client.Room.IsEmpty() {
		delete(h.Rooms, client.Room.ID)
		slog.Debug("room deleted because it was empty", "room_id", client.Room.ID)
	}
	close(client.Send)
	slog.Debug("client unregistered", "client_id", client.ID, "room_id", client.Room.ID)
}

// Creates or overwrites a room with the given id
func (h *Hub) CreateRoom(id string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	room := NewRoom(id)
	h.Rooms[id] = room
	slog.Debug("room created", "room_id", id)
	return room
}

// Get the room and see if it exists
func (h *Hub) GetRoom(id string) (*Room, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	room, exists := h.Rooms[id]
	return room, exists
}
