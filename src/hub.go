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
			if connections, ok := hub.clients[client.group]; ok {
				if _, exists := connections[client]; exists {
					// Remove the client and close their channel
					delete(connections, client)
					close(client.send)

					// Broadcast departure
					// We do this here so it's guaranteed to fire exactly once per disconnect
					hub.broadcastUserLeft(client)

					// Cleanup group if empty
					// len(connections) works here because connections points to the same underlying memory as the map(a reference type) stored in the hub.clients registry
					// delete(connections, client) modified the actual map. Because connections and hub.clients[client.group] point to the same thing, the change is "live."
					if len(connections) == 0 {
						delete(hub.clients, client.group)
						hub.redis.Unsubscribe(client.group)
						slog.Info("Board empty. Unsubscribed from Redis.", "group", client.group)
					}
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

func (hub *Hub) broadcastUserLeft(c *Client) {
	userClosingEvent := &UserClosingEvent{
		By:    c.id,
		Group: c.group,
		Xid:   c.xid,
	}
	// We call Handle(nil, ...) because usually Handle(c, ...)
	// is for events coming FROM a specific client.
	// Here, the client is already gone.
	userClosingEvent.Handle(nil, hub)
}
