package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// go test -bench="."

// Constants for the simulation
const (
	BenchBoardID = "benchmark-board"
	NumUsers     = 50
	NumMessages  = 100
	NumComments  = 20
	NumColumns   = 5
)

// setupBenchmarkData populates Redis using the actual application methods
func setupBenchmarkData(c *RedisConnector) {
	// 1. Clear DB for a clean test
	c.client.FlushDB(c.ctx)

	// 2. Create Board & Columns
	board := &Board{
		Id:     BenchBoardID,
		Name:   "Bench Board",
		Team:   "QA Team",
		Owner:  "user0",
		Status: 0, // InProgress
		Mask:   false,
		Lock:   false,
	}

	cols := make([]*BoardColumn, NumColumns)
	for i := 0; i < NumColumns; i++ {
		cols[i] = &BoardColumn{
			Id:        fmt.Sprintf("col%d", i),
			Text:      fmt.Sprintf("Column %d", i),
			Color:     "#ffffff",
			Position:  i,
			IsDefault: i == 0,
		}
	}

	if !c.CreateBoard(board, cols) {
		panic("Failed to create board via helper method")
	}

	// 3. Create Users
	for i := 0; i < NumUsers; i++ {
		u := &User{
			Id:       fmt.Sprintf("user%d", i),
			Xid:      fmt.Sprintf("xid-%d", i),
			Nickname: fmt.Sprintf("Tester %d", i),
		}
		if !c.CommitUserPresence(BenchBoardID, u, true) {
			panic("Failed to create user via helper method")
		}
	}

	// 4. Create Messages & Likes
	for i := 0; i < NumMessages; i++ {
		msgID := fmt.Sprintf("msg%d", i)
		msg := &Message{
			Id:         msgID,
			By:         "user0",
			ByXid:      "xid-0",
			ByNickname: "Tester 0",
			Group:      BenchBoardID,
			Content:    "Hello World",
			Category:   "col0",
		}

		// Save as New Message
		if !c.Save(msg, AsNewMessage) {
			panic("Failed to save message via helper method")
		}

		// Add some likes (User0, User1, User2 like every message)
		for u := 0; u < 3; u++ {
			likerID := fmt.Sprintf("user%d", u)
			if !c.Like(msgID, likerID, true) {
				// It's okay if this fails on duplicate likes, but shouldn't fail on fresh data
				slog.Warn("Like failed", "msg", msgID, "user", likerID)
			}
		}
	}

	// 5. Create Comments
	for i := 0; i < NumComments; i++ {
		cmtID := fmt.Sprintf("cmt%d", i)
		cmt := &Message{
			Id:         cmtID,
			By:         "user1",
			ByXid:      "xid-1",
			ByNickname: "Tester 1",
			Group:      BenchBoardID,
			Content:    "This is a comment",
			ParentId:   "msg0", // Assuming msg0 exists
		}

		// Save as New Comment
		if !c.Save(cmt, AsNewComment) {
			panic("Failed to save comment via helper method")
		}
	}
}

// initRedis connects to local Redis and sets up the connector
func initRedis() *RedisConnector {
	// Simple slog handler for tests
	opts := &slog.HandlerOptions{Level: slog.LevelError}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	r := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	return &RedisConnector{
		ctx:        context.Background(),
		client:     r,
		timeToLive: 24 * time.Hour, // Important: Set TTL as your methods use it
	}
}

// --- BENCHMARK: OLD SEQUENTIAL APPROACH ---
// go test -bench=BenchmarkFetch_Sequential -benchmem -v
// go test -bench=BenchmarkFetch_Sequential -benchmem -benchtime=5s
func BenchmarkFetch_Sequential(b *testing.B) {
	conn := initRedis()
	setupBenchmarkData(conn) // Passing conn, not just client
	// currentUserID := "user1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Mimic your original N+1 logic here if you still have the old methods
		// or leave this empty if you only want to test the new one.
		// For comparison, you would ideally call the old "Broadcast" logic here.

		// 1. Get Board
		board, _ := conn.GetBoard(BenchBoardID)
		if board == nil {
			continue
		}

		// 2. Get Columns
		conn.GetBoardColumns(BenchBoardID)

		// 3. Get Users
		conn.GetUsersPresence(BenchBoardID)

		// 4. Get Messages
		msgs, _ := conn.GetMessages(BenchBoardID)

		// 5. Get Comments
		conn.GetComments(BenchBoardID)

		// 6. Get Likes (Sequential to the rest)
		msgIDs := make([]string, len(msgs))
		for j, m := range msgs {
			msgIDs[j] = m.Id
		}
	}
}

// --- BENCHMARK: NEW PIPELINED APPROACH ---
// go test -bench=BenchmarkFetch_Pipelined -benchmem -v
// go test -bench=BenchmarkFetch_Pipelined -benchmem -benchtime=5s
func BenchmarkFetch_Pipelined(b *testing.B) {
	conn := initRedis()
	setupBenchmarkData(conn) // Passing conn, not just client
	// currentUserID := "user1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ok := conn.GetBoardAggregatedData(BenchBoardID)
		if !ok {
			b.Fatal("GetBoardAggregatedData failed! Run with -v to see details.")
		}
	}
}
