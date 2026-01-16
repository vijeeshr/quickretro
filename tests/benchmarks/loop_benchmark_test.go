package main

import (
	"fmt"
	"testing"
)

// go test -bench="."
// go test -bench="." -benchmem -v
type Response struct {
	ID   int
	Data [10]int // Adding some payload
}

// type Client struct {
// 	id   int
// 	send chan any
// }

// go test -bench=BenchmarkLoopApproach1 -benchmem -v
// go test -bench=BenchmarkLoopApproach1 -benchmem -benchtime=5s
func BenchmarkLoopApproach1(b *testing.B) {
	clients := make([]*Client, 1000)
	for i := 0; i < 1000; i++ {

		clients[i] = &Client{id: fmt.Sprintf("id%d", i), send: make(chan any, 1)}
	}
	regResponse := Response{ID: 1}
	joinResp := Response{ID: 2}
	senderID := "id500"

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, client := range clients {
			if client.id == senderID {
				select {
				case client.send <- regResponse:
				default:
				}
				continue
			}
			select {
			case client.send <- joinResp:
			default:
			}
		}
	}
}

// go test -bench=BenchmarkLoopApproach2 -benchmem -v
// go test -bench=BenchmarkLoopApproach2 -benchmem -benchtime=5s
func BenchmarkLoopApproach2(b *testing.B) {
	clients := make([]*Client, 1000)
	for i := 0; i < 1000; i++ {
		clients[i] = &Client{id: fmt.Sprintf("id%d", i), send: make(chan any, 1)}
	}
	regResponse := Response{ID: 1}
	joinResp := Response{ID: 2}
	senderID := "id500"

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, client := range clients {
			var payload any
			if client.id == senderID {
				payload = regResponse
			} else {
				payload = joinResp
			}
			select {
			case client.send <- payload:
			default:
			}
		}
	}
}

// BenchmarkLoopApproach1-16          75273             14160 ns/op               0 B/op          0 allocs/op
// BenchmarkLoopApproach1-16          76764             15276 ns/op               0 B/op          0 allocs/op
// BenchmarkLoopApproach1-16          81417             14648 ns/op               0 B/op          0 allocs/op

// BenchmarkLoopApproach2-16          74014             15402 ns/op               0 B/op          0 allocs/op
// BenchmarkLoopApproach2-16          80476             15761 ns/op               0 B/op          0 allocs/op
// BenchmarkLoopApproach2-16          71023             15200 ns/op               0 B/op          0 allocs/op
