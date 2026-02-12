package main

import (
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	// maxMessageSize = 1024 //512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// slog.Info("Origin information", "Origin", origin)

		// switch origin {
		// case "http://localhost:8080", "https://localhost:8080", "http://localhost:5173",
		// 	"https://localhost", "https://quickretro.app", "https://demo.quickretro.app":
		// 	return true
		// default:
		// 	return false
		// }
		return slices.Contains(config.Server.AllowedOrigins, origin)
	},
}

type Client struct {
	hub   *Hub
	conn  *websocket.Conn
	send  chan any
	id    string // This is the user uuid
	xid   string // The is the externally exposed uuid of the user
	group string // This can be a board/room
}

func (c *Client) read() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(config.Websocket.MaxMessageSizeBytes)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		// slog.Debug("Pong", "from", c.id)
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		// Read from socket and parse
		var event Event
		err := c.conn.ReadJSON(&event)
		if err != nil {
			slog.Error("Disconnecting", "err", err, "user", c.id)
			break
		}

		// Always overwrite client-sent group/by/xid (never trust UI)
		event.Group = c.group
		event.By = c.id
		event.Xid = c.xid

		event.Handle(c.hub)
	}
}

func (c *Client) write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		// delete(c.hub.clients, c) // Should this be deleted?. Race condition?
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(message); err != nil {
				slog.Error("Error when writing to socket", "err", err, "user", c.id)
				return // return or break?
			}
		case <-ticker.C:
			// slog.Debug("Ping", "To", c.id)
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("Error writing PingMessage to socket", "err", err, "user", c.id)
				return
			}
		}
	}
}

func handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Grab values from request. Validate etc
	board, ok := mux.Vars(r)["board"]
	if !ok || board == "" || len(board) > MaxIdSizeBytes {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, ok := mux.Vars(r)["user"]
	if !ok || user == "" || len(user) > MaxIdSizeBytes {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	xid := r.URL.Query().Get("xid")
	if xid == "" || len(xid) > MaxIdSizeBytes {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// if !hub.redis.BoardExists(board) {
	// 	w.WriteHeader(http.StatusNotFound)
	// 	return
	// }

	// If board doesn't exist, upgrade and close immediately with a reason
	if !hub.redis.BoardExists(board) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("Board not found", "board", board)
			return
		}
		// Send close control frame with code + reason
		msg := websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "BOARDNOTFOUND")
		_ = conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(time.Second))
		conn.Close()
		return
	}

	slog.Info("Connecting", "board", board, "user", user)
	// Upgrade http request to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error when upgrading to websocket", "err", err)
		return
	}

	// Represent the websocket connection as a "Client".
	client := &Client{id: user, xid: xid, group: board, conn: conn, send: make(chan any, 256), hub: hub}

	// Register the connection/client with the Hub
	client.hub.register <- client
	client.hub.redis.Subscribe(board)

	go client.read()
	go client.write()
}
