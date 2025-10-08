package signaling

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

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
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{
				"*",
			},
		})
		if err != nil {
			slog.Debug("upgrade to websocket", "error", err)
			http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
			return
		}
		clientCtx, cancel := context.WithTimeout(context.Background(), time.Minute*10)

		defer func() {
			slog.Debug("closing connection inside server.go")
			c.CloseNow() // this is a noop if already closed
			cancel()
		}()
		// max of 10 mins connection time
		client := &Client{
			ID:     shortuuid.New(),
			Hub:    hub,
			Room:   room,
			Conn:   c,
			Send:   make(chan []byte, 256),
			IsHost: isHost,
			Ctx:    clientCtx,
		}

		if err := room.AddClient(client); err != nil {
			slog.Debug("add client to room", "error", err)
			client.SendMessage(&ErrorMessage{
				Type:    "error",
				Message: err.Error(),
			})
			return
		}

		client.SendMessage(&RoomMetaMessage{
			Type:   MessageTypeRoomMeta,
			RoomId: room.ID,
		})

		slog.Debug("client connected", "client_id", client.ID, "room_id", room.ID, "is_host", client.IsHost)
		client.ReadWriteWs(clientCtx) // blocking
		slog.Debug("client disconnected", "client_id", client.ID, "room_id", room.ID)
		room.RemoveClient(client)
	})
}
