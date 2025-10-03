package signaling

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/coder/websocket"
	"github.com/lithammer/shortuuid/v4"
)

func HandleSignalServer(ctx context.Context, mux *http.ServeMux, hub *Hub) {

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("ws connection attempt", "remote_addr", r.RemoteAddr)
		roomID := r.URL.Query().Get("roomId")
		role := r.URL.Query().Get("role")

		if role != "host" && role != "client" {
			http.Error(w, "role must be 'host' or 'client'", http.StatusBadRequest)
			return
		}

		if roomID == "" && role == "client" {
			http.Error(w, "roomId is required", http.StatusBadRequest)
			return
		}

		isHost := role == "host"

		if isHost {
			// create and set the room id
			roomID = strings.ToUpper(shortuuid.New()[0:6]) // 6 char room id i.e. "AB12CD"
			hub.CreateRoom(roomID)
		}

		room, roomExists := hub.GetRoom(roomID)
		if !roomExists {
			slog.Error("room does not exist", "room_id", roomID)
			http.Error(w, "Could not find room, try again", http.StatusInternalServerError)
			return
		}
		// upgrade connection to websocket
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
		if err != nil {
			http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
			return
		}
		defer c.CloseNow()

		client := &Client{
			ID:     shortuuid.New(),
			Hub:    hub,
			Room:   room,
			Conn:   c,
			Send:   make(chan []byte, 256),
			IsHost: isHost,
		}

		hub.Register <- client

		go client.WritePump(ctx)
		go client.ReadPump(ctx)
		slog.Debug("client connected", "client_id", client.ID, "room_id", room.ID, "is_host", client.IsHost)

	})
}
