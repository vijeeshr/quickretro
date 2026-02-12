package main

import (
	"fmt"
	"testing"
)

// go test -bench="."
// go test -bench="." -benchmem -v

// go test -bench=BenchmarkFmtSprintfBoardKey -benchmem -v
// go test -bench=BenchmarkFmtSprintfBoardKey -benchmem -benchtime=5s
func BenchmarkFmtSprintfBoardKey(b *testing.B) {
	boardId := "abcd1234"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("board:%s", boardId)
	}
}

// go test -bench=BenchmarkBoardKeyConcat -benchmem -v
// go test -bench=BenchmarkBoardKeyConcat -benchmem -benchtime=5s
func BenchmarkBoardKeyConcat(b *testing.B) {
	boardId := "abcd1234"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = boardKey(boardId)
	}
}

// go test -bench=BenchmarkFmtSprintfBoardUserKey -benchmem -v
// go test -bench=BenchmarkFmtSprintfBoardUserKey -benchmem -benchtime=5s
func BenchmarkFmtSprintfBoardUserKey(b *testing.B) {
	boardId := "board1"
	userId := "user1"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("board:user:%s:%s", boardId, userId)
	}
}

// go test -bench=BenchmarkBoardUserKeyConcat -benchmem -v
// go test -bench=BenchmarkBoardUserKeyConcat -benchmem -benchtime=5s
func BenchmarkBoardUserKeyConcat(b *testing.B) {
	boardId := "board1"
	userId := "user1"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = boardUserKey(boardId, userId)
	}
}

// Results
// BenchmarkFmtSprintfBoardKey-16           8756272               141.4 ns/op            16 B/op          1 allocs/op
// BenchmarkBoardKeyConcat-16              41217918                29.61 ns/op            0 B/op          0 allocs/op
// BenchmarkFmtSprintfBoardUserKey-16       6431750               187.3 ns/op            24 B/op          1 allocs/op
// BenchmarkBoardUserKeyConcat-16          24346632                50.35 ns/op            0 B/op          0 allocs/op
