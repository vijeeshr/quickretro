package main

import (
	"encoding/json"
	"log"
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
			hub.clients[client.group][client] = true // Insert of Update
			// hub.clients[client] = true
		case client := <-hub.unregister:
			// Delete the client and close the client's sending channel
			// Also delete the group/board/room when there are no clients attached to it.
			if _, ok := hub.clients[client.group]; ok {
				delete(hub.clients[client.group], client)
				close(client.send)
				// Delete group/board/room if not clients exist
				if len(hub.clients[client.group]) == 0 {
					delete(hub.clients, client.group)
				}
			}
		case broadcast := <-hub.redis.subscriber.Channel():
			var args BroadcastArgs
			if err := json.Unmarshal([]byte(broadcast.Payload), &args); err != nil {
				log.Printf("error unmarshalling to BroadcastArgs from redis channel: %v", err)
			}
			args.Event.Broadcast(args.Message, hub)
		}
	}
}

// func (hub *Hub) executeMasking(i Incoming) {
// 	// Validate if command came from board owner
// 	if board, ok := hub.redis.GetBoard(i.Group); ok && board.Owner == i.By {
// 		// Save and broadcast
// 		if ok := hub.redis.SaveBoard(board, i.Type); ok {
// 			res := MaskResponse{Type: i.Type}
// 			// broadcast
// 			clients := hub.clients[i.Group]
// 			for client := range clients {
// 				select {
// 				case client.send <- res:
// 				default:
// 					delete(hub.clients[client.group], client)
// 					close(client.send)
// 					// Todo: Should the group be deleted from here if there are no more clients remaining? This is a place to broadcast though.
// 					// delete(hub.clients, client)
// 					// close(client.send)
// 				}
// 			}
// 		}
// 	}
// }
