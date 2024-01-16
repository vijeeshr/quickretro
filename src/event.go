package main

import (
	"encoding/json"
	"log"
)

type Event struct {
	Type    string          `json:"typ"` // Values can be one of "reg", "msg", "del", "like".
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
	case "present":
		payload.(*PresentEvent).Handle(e, h)
	case "msg":
		payload.(*MessageEvent).Handle(e, h)
	case "like":
		payload.(*LikeMessageEvent).Handle(e, h)
	case "del":
		payload.(*DeleteMessageEvent).Handle(e, h)
	}
}

// Broadcast event. This is executed when Redis pubsub sends message/data. Hub gets the message first, which is forwarded here.
// Note "reg" isn't inculded here. That's because it returns back to the same webcocket connection. No redis pubsub broadcast is used here.
func (e *Event) Broadcast(m *Message, h *Hub) {
	payload := e.ParsePayload()
	if payload == nil {
		return
	}
	// Call individual broadcasters
	switch e.Type {
	case "mask":
		payload.(*MaskEvent).Broadcast(h)
	case "present":
		payload.(*PresentEvent).Broadcast(h)
	case "msg":
		payload.(*MessageEvent).Broadcast(m, h)
	case "like":
		payload.(*LikeMessageEvent).Broadcast(m, h)
	case "del":
		payload.(*DeleteMessageEvent).Broadcast(m, h)
	}
}

func (e *Event) ParsePayload() interface{} {
	// Todo: Check allocations.
	payloadMap := map[string]interface{}{
		"mask":    &MaskEvent{},
		"present": &PresentEvent{},
		"msg":     &MessageEvent{},
		"like":    &LikeMessageEvent{},
		"del":     &DeleteMessageEvent{},
	}
	payload, ok := payloadMap[e.Type]
	if !ok {
		log.Println("unsupported command type")
		return nil
	}
	if err := json.Unmarshal(e.Payload, payload); err != nil {
		log.Printf("error unmarshalling event payload: %v", err)
		return nil
	}
	log.Printf("event payload after full Unmarshalling: %v", payload)
	return payload
}
