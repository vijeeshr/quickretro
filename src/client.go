package main

import (
	"encoding/json"
	"fmt"
	"log"
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
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { log.Println("Pong"); c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		// Read from socket and parse
		var event Event
		err := c.conn.ReadJSON(&event)
		if err != nil {
			log.Println(err)
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				removePresence := &RemoveUserPresenceEvent{By: c.id, Group: c.group}
				removePresence.Handle(c.hub)
				break
			}
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// RegisterEvent should happen just once per connection. Ideally, the first call using the connection.
		// Attach the userid to the client. The Client immediately sends the id after establishing connection.
		if event.Type == "reg" {
			var regEvent RegisterEvent
			if err = json.Unmarshal(event.Payload, &regEvent); err != nil {
				log.Printf("error parsing RegisterEvent: %v", err)
				break
			}
			// regEvent.Register(c)
			regEvent.Handle(c)
			continue
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
				log.Println("Error when writing to socket conn", err)
				return // return or break?
			}
		case <-ticker.C:
			log.Println("Ping")
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Grab values from request. Validate etc
	id, ok := mux.Vars(r)["id"]
	if !ok || id == "" {
		// If board is not passed, return as Bad request.
		fmt.Println("board not passed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !hub.redis.BoardExists(id) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Println("Initiating websocket connection")
	// Upgrade http request to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Represent the websocket connection as a "Client". "id" is set later when the browser sends a "reg" event.
	client := &Client{id: "", group: id, conn: conn, send: make(chan interface{}, 256), hub: hub}

	// Register the connection/client with the Hub
	client.hub.register <- client
	client.hub.redis.Subscribe(id) // ToDo: This subscribes to the same redis channels everytime a request comes for the same board/group. Check the impact.

	go client.read()
	go client.write()
}
