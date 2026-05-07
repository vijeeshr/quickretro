package main

import (
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// writeItem represents a single payload to be written to a client's WebSocket connection.
// A nil payload signals the worker to send a PingMessage instead of JSON.
type writeItem struct {
	client  *Client
	payload any
}

// BoardWriter manages a pool of worker goroutines that write to all WebSocket
// connections for a single board. This replaces per-client write goroutines
// with a shared pool, reducing goroutine count and providing centralized
// slow-client eviction with configurable write deadlines.
type BoardWriter struct {
	board         string
	hub           *Hub
	writeCh       chan writeItem
	done          chan struct{}
	writeDeadline time.Duration
	workerCount   int
	closeOnce     sync.Once
}

func newBoardWriter(board string, hub *Hub, bufSize int, workerCount int, deadline time.Duration) *BoardWriter {
	return &BoardWriter{
		board:         board,
		hub:           hub,
		writeCh:       make(chan writeItem, bufSize),
		done:          make(chan struct{}),
		writeDeadline: deadline,
		workerCount:   workerCount,
	}
}

// run launches the worker pool and a single ping loop goroutine.
func (bw *BoardWriter) run() {
	for i := 0; i < bw.workerCount; i++ {
		go bw.worker(i)
	}
	go bw.pingLoop()
}

// worker consumes writeItems from the shared channel and writes to client sockets.
func (bw *BoardWriter) worker(id int) {
	slog.Debug("Board writer worker started", "board", bw.board, "worker", id)
	for {
		select {
		case item, ok := <-bw.writeCh:
			if !ok {
				slog.Debug("Board writer worker stopping (channel closed)", "board", bw.board, "worker", id)
				return
			}
			bw.writeToClient(item)
		case <-bw.done:
			slog.Debug("Board writer worker stopping (done signal)", "board", bw.board, "worker", id)
			return
		}
	}
}

// writeToClient performs a mutex-protected write to a single client's socket.
// gorilla/websocket only supports one concurrent writer per connection,
// so the per-client writeMu ensures safety when workerCount > 1.
func (bw *BoardWriter) writeToClient(item writeItem) {
	c := item.client
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	c.conn.SetWriteDeadline(time.Now().Add(bw.writeDeadline))

	var err error
	if item.payload == nil {
		// Ping message
		err = c.conn.WriteMessage(websocket.PingMessage, nil)
		if err != nil {
			slog.Warn("Ping failed, evicting client", "user", c.id, "board", bw.board, "err", err)
			c.hub.unregister <- c
		}
	} else {
		err = c.conn.WriteJSON(item.payload)
		if err != nil {
			slog.Warn("Write failed, evicting client", "user", c.id, "board", bw.board, "err", err)
			c.hub.unregister <- c
		}
	}
}

// pingLoop runs a single ticker that enqueues ping items for all clients
// in this board. Workers pick these up and send PingMessage via writeToClient.
func (bw *BoardWriter) pingLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			clients := bw.hub.clients[bw.board]
			for client := range clients {
				select {
				case bw.writeCh <- writeItem{client: client, payload: nil}:
				default:
					// Queue full — client will miss this ping.
					// The client's pong timeout will eventually catch this.
					slog.Debug("Ping queue full, skipping client", "user", client.id, "board", bw.board)
				}
			}
		case <-bw.done:
			return
		}
	}
}

// enqueue sends a payload to the write channel for a specific client.
// If the channel is full, the client is evicted (matching legacy behavior
// where a full client.send channel triggers unregister).
func (bw *BoardWriter) enqueue(client *Client, payload any) {
	select {
	case bw.writeCh <- writeItem{client: client, payload: payload}:
	default:
		slog.Warn("Board write channel full, evicting client", "user", client.id, "board", bw.board)
		client.hub.unregister <- client
	}
}

// stop signals all workers and the ping loop to exit.
func (bw *BoardWriter) stop() {
	bw.closeOnce.Do(func() {
		close(bw.done)
		slog.Debug("Board writer stopped", "board", bw.board)
	})
}
