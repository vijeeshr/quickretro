package main

import (
	"encoding/json"
	"log/slog"
)

type Hub struct {
	clients    map[string]map[*Client]bool // Board-wise clients. Board is like a typical "room".
	register   chan *Client
	unregister chan *Client
	redis      *RedisConnector
}

func newHub(r *RedisConnector) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		redis:      r,
	}
}

func (hub *Hub) run() {
	for {
		select {
		case client := <-hub.register:
			// Create group/board/room if it doesn't exist
			if _, ok := hub.clients[client.group]; !ok {
				hub.clients[client.group] = make(map[*Client]bool)
			}
			hub.clients[client.group][client] = true // Insert or Update
			// hub.clients[client] = true
		case client := <-hub.unregister:
			// Delete the client and close the client's sending channel
			// Also delete the group/board/room when there are no clients attached to it.
			if _, ok := hub.clients[client.group]; ok {
				delete(hub.clients[client.group], client)
				close(client.send)
				// Delete group/board/room if no clients exist
				if len(hub.clients[client.group]) == 0 {
					delete(hub.clients, client.group)
				}
			}
		case broadcast := <-hub.redis.subscriber.Channel():
			var args BroadcastArgs
			if err := json.Unmarshal([]byte(broadcast.Payload), &args); err != nil {
				slog.Error("Error unmarshalling to BroadcastArgs from redis channel in hub", "details", err.Error(), "payload", broadcast.Payload)
			}
			args.Event.Broadcast(args.Message, hub)
		}
	}
}
