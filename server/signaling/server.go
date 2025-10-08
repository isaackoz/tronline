package signaling

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/lithammer/shortuuid/v4"
)

func HandleSignalServer(rootCtx context.Context, mux *http.ServeMux, hub *Hub) {

	/*
		Error code reference
		3000 = unknown server error
		3001 = invalid role
		3002 = missing roomId when role=client
		3003 = room id collision (host) or could not create room (host)
		3004 = room does not exist (client)
		3005 = could not add client to room (room full etc)

	*/
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("ws connection attempt", "remote_addr", r.RemoteAddr)
		// upgrade connection to websocket
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{
				"*",
			},
		})
		if err != nil {
			slog.Debug("upgrade to websocket", "error", err)
			http.Error(w, "Could not open websocket connection", http.StatusInternalServerError)
			return
		}

		clientCtx, cancel := context.WithTimeout(context.Background(), time.Minute*10)

		defer func() {
			slog.Debug("closing connection inside server.go")
			c.CloseNow() // this is a noop if already closed
			cancel()
		}()

		roomID := r.URL.Query().Get("roomId")
		role := r.URL.Query().Get("role")
		if role != "host" && role != "client" {
			c.Close(3001, "role must be 'host' or 'client'")
			return
		}

		if roomID == "" && role == "client" {
			c.Close(3002, "roomId is required when role is 'client'")
			return
		}

		isHost := role == "host"

		if isHost {
			// create and set the room id
			roomID = strings.ToUpper(shortuuid.New()[0:6]) // 6 char room id i.e. "AB12CD"
			_, roomExists := hub.GetRoom(roomID)
			if roomExists {
				// the odds are pretty damn low, but if it does exist, just close the connection and let them try again
				slog.Error("room id collision", "room_id", roomID)
				c.Close(3003, "room id collision (rare!), please try again")
				return
			}
			room := hub.CreateRoom(roomID)
			if room == nil {
				c.Close(3003, "could not create room, please try again")
				return
			}
			go hub.RunRoom(rootCtx, room) // runs the rooms logic outside this request handler
		}

		room, roomExists := hub.GetRoom(roomID)
		if !roomExists {
			slog.Debug("room does not exist", "room_id", roomID)
			c.Close(3004, "room does not exist")
			return
		}

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
			var roomErr *RoomError
			if errors.As(err, &roomErr) {
				slog.Debug("could not add client to room", "error", err, "room_id", room.ID)
				c.Close(3005, roomErr.Message)
			} else {
				c.Close(3000, "internal server error")
			}
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
