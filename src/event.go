package main

import (
	"encoding/json"
	"log/slog"
)

type Event struct {
	Type    string          `json:"typ"` // Values can be one of "reg", "msg", "del", "like", "mask", "timer", "catchng". "closing" is not initiated from UI.
	Payload json.RawMessage `json:"pyl"`
}

// Handle event
func (e *Event) Handle(h *Hub) {
	payload := e.ParsePayload()
	if payload == nil {
		return
	}
	// Call individual handlers
	switch e.Type {
	case "mask":
		payload.(*MaskEvent).Handle(e, h)
	case "lock":
		payload.(*LockEvent).Handle(e, h)
	case "reg":
		payload.(*RegisterEvent).Handle(e, h)
	case "msg":
		payload.(*MessageEvent).Handle(e, h)
	case "like":
		payload.(*LikeMessageEvent).Handle(e, h)
	case "del":
		payload.(*DeleteMessageEvent).Handle(e, h)
	case "catchng":
		payload.(*CategoryChangeEvent).Handle(e, h)
	case "timer":
		payload.(*TimerEvent).Handle(e, h)
	}
}

// Broadcast event. This is executed when Redis pubsub sends message/data. Hub gets the message first, which is forwarded here.
func (e *Event) Broadcast(m *Message, h *Hub) {
	payload := e.ParsePayload()
	if payload == nil {
		return
	}
	// Call individual broadcasters
	switch e.Type {
	case "mask":
		payload.(*MaskEvent).Broadcast(h)
	case "lock":
		payload.(*LockEvent).Broadcast(h)
	case "reg":
		payload.(*RegisterEvent).Broadcast(h)
	case "msg":
		payload.(*MessageEvent).Broadcast(m, h)
	case "like":
		payload.(*LikeMessageEvent).Broadcast(m, h)
	case "del":
		payload.(*DeleteMessageEvent).Broadcast(m, h)
	case "catchng":
		payload.(*CategoryChangeEvent).Broadcast(h)
	case "timer":
		payload.(*TimerEvent).Broadcast(h)
	case "closing":
		payload.(*UserClosingEvent).Broadcast(h)
	}
}

func (e *Event) ParsePayload() interface{} {
	// Todo: Check allocations.
	payloadMap := map[string]interface{}{
		"mask":    &MaskEvent{},
		"lock":    &LockEvent{},
		"reg":     &RegisterEvent{},
		"msg":     &MessageEvent{},
		"like":    &LikeMessageEvent{},
		"del":     &DeleteMessageEvent{},
		"catchng": &CategoryChangeEvent{},
		"timer":   &TimerEvent{},
		"closing": &UserClosingEvent{},
	}
	payload, ok := payloadMap[e.Type]
	if !ok {
		slog.Error("Unsupported command type", "commandType", e.Type)
		return nil
	}
	if err := json.Unmarshal(e.Payload, payload); err != nil {
		slog.Error("Error unmarshalling event payload", "details", err.Error())
		return nil
	}
	slog.Debug("Unmarshalled event payload", "payload", payload)
	return payload
}
