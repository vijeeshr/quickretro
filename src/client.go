package main

import (
	"log/slog"
	"net/http"
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
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		slog.Info("Origin information", "Origin", r.Header.Get("Origin"))
		return true
		// Todo: Uncomment below code
		// origin := r.Header.Get("Origin")
		// switch origin {
		// case "http://localhost:8080":
		// 	return true
		// default:
		// 	return false
		// }
	},
}

type Client struct {
	id    string // This is the user uuid
	group string // This can be a board/room
	conn  *websocket.Conn
	send  chan interface{}
	hub   *Hub
}

func (c *Client) read() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		// Cleanup subscription if there are no users left in the board
		presentUsers, ok := c.hub.redis.GetPresentUserIds(c.group)
		if ok && len(presentUsers) == 0 {
			slog.Info("No users left in board. Unsubscribing.", "board", c.group)
			c.hub.redis.Unsubscribe(c.group)
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
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
			slog.Error("Error reading from socket", "details", err.Error(), "user", c.id)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
				slog.Error("Unexpected close error when reading from socket", "details", err.Error(), "user", c.id)
			}
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
				userClosingEvent := &UserClosingEvent{By: c.id, Group: c.group}
				userClosingEvent.Handle(c.hub)
			}
			break
		}

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
				slog.Error("Error when writing to socket", "details", err.Error(), "user", c.id)
				return // return or break?
			}
		case <-ticker.C:
			// slog.Debug("Ping", "To", c.id)
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("Error writing PingMessage to socket", "details", err.Error(), "user", c.id)
				return
			}
		}
	}
}

func handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Grab values from request. Validate etc
	board, ok := mux.Vars(r)["board"]
	if !ok || board == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, ok := mux.Vars(r)["user"]
	if !ok || user == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !hub.redis.BoardExists(board) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	slog.Info("Initiating websocket connection", "board", board, "user", user)
	// Upgrade http request to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error when upgrading to websocket", "details", err.Error())
		return
	}

	// Represent the websocket connection as a "Client".
	client := &Client{id: user, group: board, conn: conn, send: make(chan interface{}, 256), hub: hub}

	// Register the connection/client with the Hub
	client.hub.register <- client
	client.hub.redis.Subscribe(board)

	go client.read()
	go client.write()
}
