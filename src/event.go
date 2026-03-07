package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime/debug"
)

type Event struct {
	Type string `json:"typ"` // Values can be one of "reg", "msg", "del", "delall", "like", "mask", "timer", "catchng". "closing" is not initiated from UI.

	// "Group", "By", "Xid" are ignored when sent from client. Each client's read goroutine overwrites them all the time.
	// This is intended for allowing json marshalling/unmarshalling for redis pubsub. With `json:"-"` those fields will loose values during pubsub.
	Group string `json:"grp"`
	By    string `json:"by"`
	Xid   string `json:"xid"`

	Payload json.RawMessage `json:"pyl"`
}

type EventHandler interface {
	Handle(e *Event, h *Hub)
	Broadcast(e *Event, m *Message, h *Hub)
}

type eventFactory func(data json.RawMessage) (EventHandler, error)

var registry = map[string]eventFactory{
	"mask":     makeFactory[MaskEvent](),
	"lock":     makeFactory[LockEvent](),
	"reg":      makeFactory[RegisterEvent](),
	"msg":      makeFactory[MessageEvent](),
	"like":     makeFactory[LikeMessageEvent](),
	"del":      makeFactory[DeleteMessageEvent](),
	"delall":   makeFactory[DeleteAllEvent](),
	"catchng":  makeFactory[CategoryChangeEvent](),
	"timer":    makeFactory[TimerEvent](),
	"colreset": makeFactory[ColumnsChangeEvent](),
	"closing":  makeFactory[UserClosingEvent](),
	"t":        makeFactory[TypedEvent](),
}

func makeFactory[T any, PT interface {
	*T
	EventHandler
}]() eventFactory {
	return func(data json.RawMessage) (EventHandler, error) {
		var target T
		if err := json.Unmarshal(data, &target); err != nil {
			return nil, err
		}
		return PT(&target), nil
	}
}

func (e *Event) GetHandler() (EventHandler, error) {
	factory, ok := registry[e.Type]
	if !ok {
		return nil, fmt.Errorf("unknown event type: %s", e.Type)
	}

	return factory(e.Payload)
}

func (e *Event) Handle(h *Hub) {
	// // Recover from any potential panics within handlers
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		slog.Error("Recovered from panic in Handle", "type", e.Type, "err", r)
	// 	}
	// }()

	handler, err := e.GetHandler()
	if err != nil {
		slog.Error("Handle failed", "type", e.Type, "err", err)
		return
	}
	handler.Handle(e, h)
}

func (e *Event) Broadcast(m *Message, h *Hub) {
	// Protect the Hub's broadcast loop from crashing
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Recovered from panic in Broadcast", "type", e.Type, "err", r, "stack", string(debug.Stack()))
		}
	}()

	handler, err := e.GetHandler()
	if err != nil {
		slog.Error("Broadcast failed", "type", e.Type, "err", err)
		return
	}
	handler.Broadcast(e, m, h)
}
