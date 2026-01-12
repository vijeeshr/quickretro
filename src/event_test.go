package main

import (
	"encoding/json"
	"testing"
)

// go test ./... -v
// go test -race ./...

// --------------------
// Test helpers / mocks
// --------------------

type mockHandler struct {
	handleCalled     bool
	broadcastCalled  bool
	panicOnBroadcast bool
}

func (m *mockHandler) Handle(e *Event, h *Hub) {
	m.handleCalled = true
}

func (m *mockHandler) Broadcast(_ *Message, _ *Hub) {
	m.broadcastCalled = true
	if m.panicOnBroadcast {
		panic("boom")
	}
}

func marshalPayload(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	return b
}

// --------------------
// GetHandler tests
// --------------------

func TestGetHandler_ValidType(t *testing.T) {
	event := &Event{
		Type: "msg",
		Payload: marshalPayload(t, MessageEvent{
			By:      "u1",
			Group:   "g1",
			Content: "hello",
		}),
	}

	h, err := event.GetHandler()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if h == nil {
		t.Fatal("expected handler, got nil")
	}

	// Verify the returned type is a pointer to MessageEvent
	msgEvent, ok := h.(*MessageEvent)
	if !ok {
		t.Fatalf("expected *MessageEvent, got %T", h)
	}

	if msgEvent.Content != "hello" {
		t.Errorf("expected payload text 'hello', got '%s'", msgEvent.Content)
	}
}

func TestGetHandler_UnknownType(t *testing.T) {
	event := &Event{
		Type:    "unknown",
		Payload: json.RawMessage(`{}`),
	}

	h, err := event.GetHandler()
	if err == nil {
		t.Fatal("expected error for unknown event type, got nil")
	}
	if h != nil {
		t.Fatal("expected nil, but got a handler for unknown event type")
	}
}

func TestGetHandler_InvalidJSON(t *testing.T) {
	event := &Event{
		Type:    "msg",
		Payload: json.RawMessage(`{"content": "missing_quote}`), // Invalid JSON //json.RawMessage(`{invalid json}`),
	}

	h, err := event.GetHandler()
	if err == nil {
		t.Fatal("error for malformed JSON payload, got nil")
	}
	if h != nil {
		t.Fatal("expected nil, but got a handler")
	}
}

// --------------------
// Handle() tests
// --------------------

func TestEvent_Handle_CallsHandler(t *testing.T) {
	mock := &mockHandler{}

	registry["test"] = func(_ json.RawMessage) (EventHandler, error) {
		return mock, nil
	}
	t.Cleanup(func() { delete(registry, "test") })

	event := &Event{
		Type:    "test",
		Payload: json.RawMessage(`{}`),
	}

	event.Handle(nil)

	if !mock.handleCalled {
		t.Fatal("expected Handle to be called")
	}
}

func TestEvent_Handle_UnknownType_DoesNotPanic(t *testing.T) {
	event := &Event{
		Type:    "nope",
		Payload: json.RawMessage(`{}`),
	}

	// Should not panic
	event.Handle(nil)
}

// func TestHandle_Safety(t *testing.T) {
// 	// Test that Handle handles errors gracefully (no type assertions to fail)
// 	e := &Event{
// 		Type:    "invalid",
// 		Payload: nil,
// 	}

// 	// Should not crash, just log an error internally via slog
// 	e.Handle(&Hub{})
// }

// --------------------
// Broadcast() tests
// --------------------

func TestEvent_Broadcast_CallsHandler(t *testing.T) {
	mock := &mockHandler{}

	registry["test"] = func(_ json.RawMessage) (EventHandler, error) {
		return mock, nil
	}
	t.Cleanup(func() { delete(registry, "test") })

	event := &Event{
		Type:    "test",
		Payload: json.RawMessage(`{}`),
	}

	event.Broadcast(nil, nil)

	if !mock.broadcastCalled {
		t.Fatal("expected Broadcast to be called")
	}
}

func TestEvent_Broadcast_RecoversFromPanic(t *testing.T) {
	mock := &mockHandler{panicOnBroadcast: true}

	registry["test"] = func(_ json.RawMessage) (EventHandler, error) {
		return mock, nil
	}
	t.Cleanup(func() { delete(registry, "test") })

	event := &Event{
		Type:    "test",
		Payload: json.RawMessage(`{}`),
	}

	// Should not panic
	event.Broadcast(nil, nil)

	if !mock.broadcastCalled {
		t.Fatal("expected Broadcast to be called even with panic")
	}
}

// // PanicEvent is used specifically to test the recovery logic
// type PanicEvent struct{}

// func (p *PanicEvent) Handle(e *Event, h *Hub)      { panic("simulated crash") }
// func (p *PanicEvent) Broadcast(m *Message, h *Hub) { panic("simulated crash") }
// func TestBroadcast_PanicRecovery(t *testing.T) {
// 	// Temporarily register a panicking event for testing
// 	registry["panic"] = makeFactory[PanicEvent]()

// 	e := &Event{
// 		Type:    "panic",
// 		Payload: json.RawMessage(`{}`),
// 	}

// 	// Execute Broadcast.
// 	// If the recovery logic fails, the test runner itself will crash.
// 	// If it succeeds, the code will log the error and return normally.
// 	defer func() {
// 		if r := recover(); r != nil {
// 			t.Errorf("Test panicked! The recover() in Broadcast failed to catch: %v", r)
// 		}
// 	}()

// 	// Passing nil for Hub/Message as they aren't needed for the panic check
// 	e.Broadcast(nil, nil)

// 	// If we reached here, the panic was caught successfully.
// }

// --------------------
// Factory behavior test
// --------------------

func TestFactory_ReturnsNewInstanceEachTime(t *testing.T) {
	event1 := &Event{
		Type: "msg",
		Payload: marshalPayload(t, MessageEvent{
			By:      "u1",
			Group:   "g1",
			Content: "one",
		}),
	}

	event2 := &Event{
		Type: "msg",
		Payload: marshalPayload(t, MessageEvent{
			By:      "u2",
			Group:   "g2",
			Content: "two",
		}),
	}

	h1, err1 := event1.GetHandler()
	h2, err2 := event2.GetHandler()

	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v %v", err1, err2)
	}

	if h1 == h2 {
		t.Fatal("expected distinct handler instances")
	}
}
