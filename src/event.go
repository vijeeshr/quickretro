package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime/debug"
)

type Event struct {
	Type string `json:"typ"` // Values can be one of "reg", "msg", "del", "delall", "like", "mask", "timer", "catchng". "closing" is not initiated from UI.

	// "Group", "By" are ignored when sent from client. Each client's read goroutine overwrites them all the time.
	// This is intended for allowing json marshalling/unmarshalling for redis pubsub. With `json:"-"` those fields will loose values during pubsub.
	Group string `json:"grp"`
	By    string `json:"by"`

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

// // Handle event
// func (e *Event) Handle(h *Hub) {
// 	payload := e.ParsePayload()
// 	if payload == nil {
// 		return
// 	}
// 	// Call individual handlers
// 	switch e.Type {
// 	case "mask":
// 		payload.(*MaskEvent).Handle(e, h)
// 	case "lock":
// 		payload.(*LockEvent).Handle(e, h)
// 	case "reg":
// 		payload.(*RegisterEvent).Handle(e, h)
// 	case "msg":
// 		payload.(*MessageEvent).Handle(e, h)
// 	case "like":
// 		payload.(*LikeMessageEvent).Handle(e, h)
// 	case "del":
// 		payload.(*DeleteMessageEvent).Handle(e, h)
// 	case "delall":
// 		payload.(*DeleteAllEvent).Handle(e, h)
// 	case "catchng":
// 		payload.(*CategoryChangeEvent).Handle(e, h)
// 	case "timer":
// 		payload.(*TimerEvent).Handle(e, h)
// 	case "colreset":
// 		payload.(*ColumnsChangeEvent).Handle(e, h)
// 	}
// }

// // Broadcast event. This is executed when Redis pubsub sends message/data. Hub gets the message first, which is forwarded here.
// func (e *Event) Broadcast(m *Message, h *Hub) {
// 	payload := e.ParsePayload()
// 	if payload == nil {
// 		return
// 	}
// 	// Call individual broadcasters
// 	switch e.Type {
// 	case "mask":
// 		payload.(*MaskEvent).Broadcast(h)
// 	case "lock":
// 		payload.(*LockEvent).Broadcast(h)
// 	case "reg":
// 		payload.(*RegisterEvent).Broadcast(h)
// 	case "msg":
// 		payload.(*MessageEvent).Broadcast(m, h)
// 	case "like":
// 		payload.(*LikeMessageEvent).Broadcast(m, h)
// 	case "del":
// 		payload.(*DeleteMessageEvent).Broadcast(m, h)
// 	case "delall":
// 		payload.(*DeleteAllEvent).Broadcast(h)
// 	case "catchng":
// 		payload.(*CategoryChangeEvent).Broadcast(h)
// 	case "timer":
// 		payload.(*TimerEvent).Broadcast(h)
// 	case "colreset":
// 		payload.(*ColumnsChangeEvent).Broadcast(h)
// 	case "closing":
// 		payload.(*UserClosingEvent).Broadcast(h)
// 	}
// }

// func (e *Event) ParsePayload() interface{} {
// 	// Todo: Check allocations.
// 	payloadMap := map[string]interface{}{
// 		"mask":     &MaskEvent{},
// 		"lock":     &LockEvent{},
// 		"reg":      &RegisterEvent{},
// 		"msg":      &MessageEvent{},
// 		"like":     &LikeMessageEvent{},
// 		"del":      &DeleteMessageEvent{},
// 		"delall":   &DeleteAllEvent{},
// 		"catchng":  &CategoryChangeEvent{},
// 		"timer":    &TimerEvent{},
// 		"colreset": &ColumnsChangeEvent{},
// 		"closing":  &UserClosingEvent{},
// 	}
// 	payload, ok := payloadMap[e.Type]
// 	if !ok {
// 		slog.Error("Unsupported command type", "commandType", e.Type)
// 		return nil
// 	}
// 	if err := json.Unmarshal(e.Payload, payload); err != nil {
// 		slog.Error("Error unmarshalling event payload", "details", err.Error())
// 		return nil
// 	}
// 	slog.Debug("Unmarshalled event payload", "payload", payload)
// 	return payload
// }
