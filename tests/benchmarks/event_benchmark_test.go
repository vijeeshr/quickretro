package main

import (
	"encoding/json"
	"testing"
)

// ---- Test helpers ----
func newTestEvent(eventType string, payload any) *Event {
	b, _ := json.Marshal(payload)
	return &Event{
		Type:    eventType,
		Payload: b,
	}
}

// ---- Old implementation (inline map) ----
func (e *Event) parsePayloadOld() any {
	payloadMap := map[string]any{
		"mask":     &MaskEvent{},
		"lock":     &LockEvent{},
		"reg":      &RegisterEvent{},
		"msg":      &MessageEvent{},
		"like":     &LikeMessageEvent{},
		"del":      &DeleteMessageEvent{},
		"delall":   &DeleteAllEvent{},
		"catchng":  &CategoryChangeEvent{},
		"timer":    &TimerEvent{},
		"colreset": &ColumnsChangeEvent{},
		"closing":  &UserClosingEvent{},
	}

	payload, ok := payloadMap[e.Type]
	if !ok {
		return nil
	}

	if err := json.Unmarshal(e.Payload, payload); err != nil {
		return nil
	}

	return payload
}

// ---- New implementation (factory map) ----
// var payloadFactories = map[string]func() any{
// 	"mask":     func() any { return &MaskEvent{} },
// 	"lock":     func() any { return &LockEvent{} },
// 	"reg":      func() any { return &RegisterEvent{} },
// 	"msg":      func() any { return &MessageEvent{} },
// 	"like":     func() any { return &LikeMessageEvent{} },
// 	"del":      func() any { return &DeleteMessageEvent{} },
// 	"delall":   func() any { return &DeleteAllEvent{} },
// 	"catchng":  func() any { return &CategoryChangeEvent{} },
// 	"timer":    func() any { return &TimerEvent{} },
// 	"colreset": func() any { return &ColumnsChangeEvent{} },
// 	"closing":  func() any { return &UserClosingEvent{} },
// }

// func (e *Event) parsePayloadNew() any {
// 	factory, ok := payloadFactories[e.Type]
// 	if !ok {
// 		return nil
// 	}

// 	payload := factory()

// 	if err := json.Unmarshal(e.Payload, payload); err != nil {
// 		return nil
// 	}

// 	return payload
// }

var payloadFactories = map[string]eventFactory{
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

func (e *Event) parsePayloadNew() any {

	// if e.handler != nil {
	// 	return e.handler
	// }

	factory, ok := payloadFactories[e.Type]
	if !ok {
		return nil
	}

	h, err := factory(e.Payload)
	if err != nil {
		return nil
	}

	// e.handler = h
	return h
}

// go test -bench="."
// go test -bench="." -benchmem -v
// go test -bench=BenchmarkParsePayload_Old -benchmem -v
// go test -bench=BenchmarkParsePayload_Old -benchmem -benchtime=5s
func BenchmarkParsePayload_Old(b *testing.B) {
	event := newTestEvent("msg", MessageEvent{
		By:      "user1",
		Group:   "board1",
		Content: "Hello world",
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = event.parsePayloadOld()
	}
}

// go test -bench="."
// go test -bench="." -benchmem -v
// go test -bench=BenchmarkParsePayload_New -benchmem -v
// go test -bench=BenchmarkParsePayload_New -benchmem -benchtime=5s
func BenchmarkParsePayload_New(b *testing.B) {
	event := newTestEvent("msg", MessageEvent{
		By:      "user1",
		Group:   "board1",
		Content: "Hello world",
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = event.parsePayloadNew()
	}
}

// go test -bench="." // Windows
// go test -bench="." -benchmem -v
// go test -bench=BenchmarkParsePayload_New_Parallel -benchmem -v
// go test -bench=BenchmarkParsePayload_New_Parallel -benchmem -benchtime=5s
func BenchmarkParsePayload_New_Parallel(b *testing.B) {
	event := newTestEvent("msg", MessageEvent{
		By:      "user1",
		Group:   "board1",
		Content: "Hello world",
	})

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = event.parsePayloadNew()
		}
	})
}

// BenchmarkParsePayload_Old-16                      173803              5904 ns/op            1600 B/op         21 allocs/op
// BenchmarkParsePayload_New-16                      273061              3966 ns/op             392 B/op          8 allocs/op
// BenchmarkParsePayload_New_Parallel-16            1464650               805.2 ns/op           392 B/op          8 allocs/op
