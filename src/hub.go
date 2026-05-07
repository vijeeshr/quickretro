package main

import (
	"encoding/json"
	"log/slog"
	"time"
)

type Hub struct {
	clients    map[string]map[*Client]bool // Board-wise clients. Board is like a typical "room".
	writers    map[string]*BoardWriter     // Board-wise writers (only used when board_writer is enabled)
	register   chan *Client
	unregister chan *Client
	redis      *RedisConnector

	// Parsed board writer config (set once at startup, read-only after)
	boardWriterDeadline time.Duration
}

func newHub(r *RedisConnector) *Hub {
	h := &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		redis:      r,
	}

	if config.Websocket.BoardWriter.Enabled {
		h.writers = make(map[string]*BoardWriter)
		deadline, err := parseDuration(config.Websocket.BoardWriter.WriteDeadline)
		if err != nil {
			slog.Warn("Invalid board writer write_deadline, using default 10s", "err", err)
			deadline = 10 * time.Second
		}
		h.boardWriterDeadline = deadline
	}

	return h
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

			// Start board writer if enabled and not already running for this board
			if config.Websocket.BoardWriter.Enabled {
				if _, exists := hub.writers[client.group]; !exists {
					bw := newBoardWriter(
						client.group,
						hub,
						config.Websocket.BoardWriter.WriteBufferSize,
						config.Websocket.BoardWriter.WorkerCount,
						hub.boardWriterDeadline,
					)
					hub.writers[client.group] = bw
					bw.run()
					slog.Debug("Board writer started", "board", client.group, "workers", bw.workerCount)
				}
			}
		case client := <-hub.unregister:
			if connections, ok := hub.clients[client.group]; ok {
				if _, exists := connections[client]; exists {
					// Remove the client and close their channel
					delete(connections, client)
					if client.send != nil {
						close(client.send)
					}

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

						// Stop board writer for this board
						if config.Websocket.BoardWriter.Enabled {
							if bw, exists := hub.writers[client.group]; exists {
								bw.stop()
								delete(hub.writers, client.group)
								slog.Debug("Board writer stopped", "board", client.group)
							}
						}
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

// SendToClient routes a payload to the appropriate write mechanism:
// - Board writer mode: enqueues into the board's shared write channel
// - Legacy mode: sends directly to the client's per-client send channel
func (hub *Hub) SendToClient(client *Client, payload any) {
	if config.Websocket.BoardWriter.Enabled {
		if writer, ok := hub.writers[client.group]; ok {
			writer.enqueue(client, payload)
			return
		}
	}
	// Legacy: direct channel send
	select {
	case client.send <- payload:
	default:
		hub.unregister <- client
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
