package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// go test -run TestBoardWriter -v
// go test -race -run TestBoardWriter -v

// --------------------
// Test helpers
// --------------------

// newTestWebSocketPair creates a real WebSocket server/client pair for testing.
// Returns the server-side connection, client-side connection, and a cleanup function.
func newTestWebSocketPair(t *testing.T) (serverConn *websocket.Conn, clientConn *websocket.Conn, cleanup func()) {
	t.Helper()

	var sConn *websocket.Conn
	var connReady sync.WaitGroup
	connReady.Add(1)

	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		sConn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}
		connReady.Done()
		// Keep handler alive until test cleanup
		select {}
	}))

	// Connect client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	cConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		server.Close()
		t.Fatalf("dial failed: %v", err)
	}

	connReady.Wait()

	return sConn, cConn, func() {
		cConn.Close()
		sConn.Close()
		server.Close()
	}
}

// newTestClient creates a Client backed by a real WebSocket connection for testing.
func newTestClient(t *testing.T, hub *Hub, group string, userId string) (*Client, *websocket.Conn, func()) {
	t.Helper()
	serverConn, clientConn, cleanup := newTestWebSocketPair(t)
	client := &Client{
		id:    userId,
		xid:   "x1",
		group: group,
		conn:  serverConn,
		hub:   hub,
	}
	return client, clientConn, cleanup
}

// newTestHub creates a minimal Hub for testing (no Redis).
func newTestHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		writers:    make(map[string]*BoardWriter),
		register:   make(chan *Client, 16),
		unregister: make(chan *Client, 16),
	}
}

// drainUnregister reads all pending unregister requests from the hub channel.
func drainUnregister(hub *Hub) []*Client {
	var evicted []*Client
	for {
		select {
		case c := <-hub.unregister:
			evicted = append(evicted, c)
		default:
			return evicted
		}
	}
}

// --------------------
// BoardWriter tests
// --------------------

func TestBoardWriter_WriteDelivery(t *testing.T) {
	hub := newTestHub()
	board := "test-board"
	hub.clients[board] = make(map[*Client]bool)

	client, clientConn, cleanup := newTestClient(t, hub, board, "user1")
	defer cleanup()
	hub.clients[board][client] = true

	bw := newBoardWriter(board, hub, 256, 1, 5*time.Second)
	bw.run()
	defer bw.stop()

	// Send a JSON payload via the board writer
	payload := map[string]string{"typ": "msg", "content": "hello"}
	bw.enqueue(client, payload)

	// Read from the client side
	clientConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var result map[string]string
	err := clientConn.ReadJSON(&result)
	if err != nil {
		t.Fatalf("expected to read message, got error: %v", err)
	}
	if result["content"] != "hello" {
		t.Errorf("expected content 'hello', got '%s'", result["content"])
	}
}

func TestBoardWriter_MultipleClients(t *testing.T) {
	hub := newTestHub()
	board := "test-board"
	hub.clients[board] = make(map[*Client]bool)

	client1, clientConn1, cleanup1 := newTestClient(t, hub, board, "user1")
	defer cleanup1()
	client2, clientConn2, cleanup2 := newTestClient(t, hub, board, "user2")
	defer cleanup2()
	hub.clients[board][client1] = true
	hub.clients[board][client2] = true

	bw := newBoardWriter(board, hub, 256, 2, 5*time.Second)
	bw.run()
	defer bw.stop()

	// Send same payload to both clients
	payload := map[string]string{"typ": "test", "data": "broadcast"}
	bw.enqueue(client1, payload)
	bw.enqueue(client2, payload)

	// Both should receive
	for i, cc := range []*websocket.Conn{clientConn1, clientConn2} {
		cc.SetReadDeadline(time.Now().Add(3 * time.Second))
		var result map[string]string
		if err := cc.ReadJSON(&result); err != nil {
			t.Fatalf("client %d: expected to read message, got error: %v", i+1, err)
		}
		if result["data"] != "broadcast" {
			t.Errorf("client %d: expected data 'broadcast', got '%s'", i+1, result["data"])
		}
	}
}

func TestBoardWriter_SlowClientEviction(t *testing.T) {
	hub := newTestHub()
	board := "test-board"
	hub.clients[board] = make(map[*Client]bool)

	client, _, cleanup := newTestClient(t, hub, board, "slow-user")
	defer cleanup()
	hub.clients[board][client] = true

	// Use an extremely short write deadline to force timeout
	bw := newBoardWriter(board, hub, 256, 1, 1*time.Millisecond)
	bw.run()
	defer bw.stop()

	// Close the client-side connection to make server writes fail
	client.conn.Close()

	// Give a moment for close to take effect
	time.Sleep(10 * time.Millisecond)

	// Now enqueue a write — it should fail and trigger eviction
	payload := map[string]string{"typ": "msg", "content": "will-fail"}
	bw.enqueue(client, payload)

	// Wait a bit for the worker to process
	time.Sleep(100 * time.Millisecond)

	evicted := drainUnregister(hub)
	if len(evicted) == 0 {
		t.Fatal("expected slow client to be evicted, but no unregister received")
	}
	if evicted[0].id != "slow-user" {
		t.Errorf("expected evicted client 'slow-user', got '%s'", evicted[0].id)
	}
}

func TestBoardWriter_WriteDeadlineEviction(t *testing.T) {
	hub := newTestHub()
	board := "test-board"
	hub.clients[board] = make(map[*Client]bool)

	client, clientConn, cleanup := newTestClient(t, hub, board, "deadline-user")
	defer cleanup()
	hub.clients[board][client] = true

	// Very tight deadline with a large buffer so enqueue never drops.
	// This ensures eviction happens due to the write deadline expiring
	// (TCP buffer full, client not reading) rather than the channel being full.
	bw := newBoardWriter(board, hub, 8192, 1, 1*time.Millisecond)
	bw.run()
	defer bw.stop()

	// Intentionally don't read from clientConn — this will cause the TCP
	// send buffer to fill, making WriteJSON block until the deadline expires.
	_ = clientConn

	// Send enough large payloads to fill the TCP buffer
	bigPayload := map[string]string{"typ": "msg", "content": strings.Repeat("x", 512)}
	for i := 0; i < 200; i++ {
		bw.enqueue(client, bigPayload)
	}

	// Wait for the worker to hit the write deadline
	time.Sleep(500 * time.Millisecond)

	evicted := drainUnregister(hub)
	if len(evicted) == 0 {
		t.Fatal("expected client to be evicted due to write deadline, but no unregister received")
	}
}

func TestBoardWriter_ChannelFullEviction(t *testing.T) {
	hub := newTestHub()
	board := "test-board"
	hub.clients[board] = make(map[*Client]bool)

	client, _, cleanup := newTestClient(t, hub, board, "full-user")
	defer cleanup()
	hub.clients[board][client] = true

	// Create a board writer with a very small buffer and don't start workers
	// so the channel fills up immediately
	bw := &BoardWriter{
		board:         board,
		hub:           hub,
		writeCh:       make(chan writeItem, 2), // tiny buffer
		done:          make(chan struct{}),
		writeDeadline: 5 * time.Second,
		workerCount:   0, // no workers — channel will fill
	}
	// Don't call bw.run() — no workers to consume

	payload := map[string]string{"typ": "msg"}

	// Fill the buffer
	bw.enqueue(client, payload)
	bw.enqueue(client, payload)

	// This third enqueue should trigger eviction (channel full)
	bw.enqueue(client, payload)

	evicted := drainUnregister(hub)
	if len(evicted) == 0 {
		t.Fatal("expected client to be evicted when write channel is full")
	}
	if evicted[0].id != "full-user" {
		t.Errorf("expected evicted client 'full-user', got '%s'", evicted[0].id)
	}
}

func TestBoardWriter_StopIsIdempotent(t *testing.T) {
	hub := newTestHub()
	bw := newBoardWriter("test-board", hub, 256, 2, 5*time.Second)
	bw.run()

	// Should not panic when called multiple times
	bw.stop()
	bw.stop()
	bw.stop()
}

func TestBoardWriter_PingEnqueue(t *testing.T) {
	hub := newTestHub()
	board := "test-board"
	hub.clients[board] = make(map[*Client]bool)

	client, clientConn, cleanup := newTestClient(t, hub, board, "ping-user")
	defer cleanup()
	hub.clients[board][client] = true

	// Set up pong handler on the client side to verify pings arrive
	pongReceived := make(chan struct{}, 1)
	clientConn.SetPingHandler(func(appData string) error {
		select {
		case pongReceived <- struct{}{}:
		default:
		}
		// Send pong back per the protocol
		return clientConn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Second))
	})

	// Start a reader on clientConn to process control frames
	go func() {
		for {
			_, _, err := clientConn.ReadMessage()
			if err != nil {
				return
			}
		}
	}()

	// Manually enqueue a ping (nil payload) and let a worker process it
	bw := newBoardWriter(board, hub, 256, 1, 5*time.Second)
	bw.run()
	defer bw.stop()

	bw.writeCh <- writeItem{client: client, payload: nil}

	select {
	case <-pongReceived:
		// Ping was received by the client — success
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for ping to be received by client")
	}
}

// --------------------
// SendToClient tests
// --------------------

func TestSendToClient_LegacyMode(t *testing.T) {
	// When board writer is disabled, SendToClient should use client.send channel
	origEnabled := config.Websocket.BoardWriter.Enabled
	config.Websocket.BoardWriter.Enabled = false
	defer func() { config.Websocket.BoardWriter.Enabled = origEnabled }()

	hub := newTestHub()
	client := &Client{
		id:    "user1",
		group: "board1",
		send:  make(chan any, 8),
		hub:   hub,
	}

	payload := map[string]string{"typ": "test"}
	hub.SendToClient(client, payload)

	select {
	case msg := <-client.send:
		result, ok := msg.(map[string]string)
		if !ok {
			t.Fatalf("expected map[string]string, got %T", msg)
		}
		if result["typ"] != "test" {
			t.Errorf("expected typ 'test', got '%s'", result["typ"])
		}
	default:
		t.Fatal("expected message on client.send channel, got nothing")
	}
}

func TestSendToClient_BoardWriterMode(t *testing.T) {
	origEnabled := config.Websocket.BoardWriter.Enabled
	config.Websocket.BoardWriter.Enabled = true
	defer func() { config.Websocket.BoardWriter.Enabled = origEnabled }()

	hub := newTestHub()
	board := "board1"
	hub.clients[board] = make(map[*Client]bool)

	client, clientConn, cleanup := newTestClient(t, hub, board, "user1")
	defer cleanup()
	hub.clients[board][client] = true

	bw := newBoardWriter(board, hub, 256, 1, 5*time.Second)
	hub.writers[board] = bw
	bw.run()
	defer bw.stop()

	payload := map[string]string{"typ": "routed"}
	hub.SendToClient(client, payload)

	// Should arrive via the board writer, not client.send
	clientConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var result map[string]string
	if err := clientConn.ReadJSON(&result); err != nil {
		t.Fatalf("expected to read routed message, got error: %v", err)
	}
	if result["typ"] != "routed" {
		t.Errorf("expected typ 'routed', got '%s'", result["typ"])
	}
}

func TestSendToClient_LegacyFullChannelEviction(t *testing.T) {
	origEnabled := config.Websocket.BoardWriter.Enabled
	config.Websocket.BoardWriter.Enabled = false
	defer func() { config.Websocket.BoardWriter.Enabled = origEnabled }()

	hub := newTestHub()
	client := &Client{
		id:    "user1",
		group: "board1",
		send:  make(chan any, 1), // tiny buffer
		hub:   hub,
	}

	payload := map[string]string{"typ": "test"}

	// Fill the channel
	hub.SendToClient(client, payload)
	// This should trigger eviction (default case)
	hub.SendToClient(client, payload)

	evicted := drainUnregister(hub)
	if len(evicted) == 0 {
		t.Fatal("expected client to be evicted when send channel is full in legacy mode")
	}
}
