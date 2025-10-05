package signaling

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/coder/websocket"
	"github.com/lithammer/shortuuid/v4"
)

func HandleSignalServer(rootCtx context.Context, mux *http.ServeMux, hub *Hub) {

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
			_, roomExists := hub.GetRoom(roomID)
			if roomExists {
				http.Error(w, "Could not create room, try again", http.StatusInternalServerError)
				return
			}
			room := hub.CreateRoom(roomID)
			if room == nil {
				http.Error(w, "Could not create room, try again", http.StatusInternalServerError)
				return
			}
			go hub.RunRoom(rootCtx, room) // runs the rooms logic outside this request handler
		}

		room, roomExists := hub.GetRoom(roomID)
		if !roomExists {
			slog.Debug("room does not exist", "room_id", roomID)
			http.Error(w, "Could not find room, try again", http.StatusInternalServerError)
			return
		}
		// upgrade connection to websocket
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
		if err != nil {
			http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
			return
		}
		defer c.CloseNow() // this is a noop if already closed

		client := &Client{
			ID:     shortuuid.New(),
			Hub:    hub,
			Room:   room,
			Conn:   c,
			Send:   make(chan []byte, 256),
			IsHost: isHost,
		}

		if err := room.AddClient(client); err != nil {
			slog.Debug("add client to room", "error", err)
			client.SendMessage(&Message{
				Type: "error",
				Data: map[string]string{"message": err.Error()},
			})
			return
		}

		client.SendMessage(&Message{
			Type: "room-meta",
			Data: map[string]any{
				"roomId": roomID,
			},
		})

		slog.Debug("client connected", "client_id", client.ID, "room_id", room.ID, "is_host", client.IsHost)
		client.ReadWriteWs(room.ctx) // blocking
		slog.Debug("client disconnected", "client_id", client.ID, "room_id", room.ID)
		room.RemoveClient(client)
	})
}
