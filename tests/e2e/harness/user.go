package harness

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

type TestUser struct {
	Id       string
	Nickname string
	Board    string
	Conn     *websocket.Conn
	Events   chan Event
	Done     chan struct{}
}

func NewUser(id, nickname, board string) *TestUser {
	return &TestUser{
		Id:       id,
		Nickname: nickname,
		Board:    board,
		Events:   make(chan Event, 100),
		Done:     make(chan struct{}),
	}
}

func (u *TestUser) Connect(baseUrl string) error {
	url := fmt.Sprintf("%s/ws/board/%s/user/%s/meet", baseUrl, u.Board, u.Id)
	// Convert http(s) to ws(s)
	if len(url) > 4 && url[:5] == "https" {
		url = "wss" + url[5:]
	} else if len(url) > 3 && url[:4] == "http" {
		url = "ws" + url[4:]
	}

	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "localhost",
	}

	conn, _, err := dialer.Dial(url, http.Header{"Origin": []string{"https://localhost"}})
	if err != nil {
		return err
	}
	u.Conn = conn

	// Start reading loop
	go func() {
		defer close(u.Done)
		for {
			_, p, err := u.Conn.ReadMessage()
			if err != nil {
				log.Printf("[User %s] Read error: %v", u.Id, err)
				return
			}

			var header struct {
				Type string `json:"typ"`
			}
			if err := json.Unmarshal(p, &header); err != nil {
				log.Printf("[User %s] Unmarshal error: %v", u.Id, err)
				continue
			}

			// log.Printf("[User %s] Received event: Type=%s Payload=%s", u.Id, header.Type, string(p))
			u.Events <- Event{
				Type: header.Type,
				Raw:  p,
			}
		}
	}()

	return nil
}

func (u *TestUser) Close() {
	if u.Conn != nil {
		u.Conn.Close()
	}
}

func (u *TestUser) SendEvent(typ string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	event := Event{
		Type:    typ,
		Payload: data,
	}
	return u.Conn.WriteJSON(event)
}

func (u *TestUser) Register() error {
	reg := RegisterEvent{
		By:         u.Id,
		ByNickname: u.Nickname,
		Xid:        "xid-" + u.Id, // Mock XID
		Group:      u.Board,
	}
	return u.SendEvent("reg", reg)
}

func (u *TestUser) SendMessage(id, content, category string) error {
	// id := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), u.Id)
	msg := MessageEvent{
		Id:         id,
		By:         u.Id,
		ByXid:      "xid-" + u.Id,
		ByNickname: u.Nickname,
		Group:      u.Board,
		Content:    content,
		Category:   category,
		ParentId:   "",
		Anonymous:  false,
	}
	// return id, u.SendEvent("msg", msg)
	return u.SendEvent("msg", msg)
}

func (u *TestUser) LikeMessage(msgId string, like bool) error {
	likeEv := LikeMessageEvent{
		MessageId: msgId,
		By:        u.Id,
		Like:      like,
	}
	return u.SendEvent("like", likeEv)
}

func (u *TestUser) LockBoard(lock bool) error {
	lockEv := LockEvent{
		By:    u.Id,
		Group: u.Board,
		Lock:  lock,
	}
	return u.SendEvent("lock", lockEv)
}

func (u *TestUser) WaitForEvent(eventType string, timeout time.Duration) (*Event, error) {
	deadline := time.Now().Add(timeout)
	for {
		timeLeft := time.Until(deadline)
		if timeLeft <= 0 {
			return nil, fmt.Errorf("timeout waiting for event %s", eventType)
		}

		select {
		case event := <-u.Events:
			if event.Type == eventType {
				return &event, nil
			}
			// log.Printf("[User %s] Skipped event %s waiting for %s", u.Id, event.Type, eventType)
		case <-time.After(timeLeft):
			return nil, fmt.Errorf("timeout waiting for event %s", eventType)
		}
	}
}

func (u *TestUser) MustWaitForEvent(t *testing.T, eventType string, target any) {
	timeout := time.After(2 * time.Second)
	for {
		select {
		case ev := <-u.Events:
			if ev.Type == eventType {
				if target != nil {
					err := json.Unmarshal(ev.Raw, target)
					require.NoError(t, err, "Failed to unmarshal event payload")
				}
				return // Success
			}
		case <-timeout:
			t.Fatalf("Timed out waiting for event: %s", eventType)
		}
	}
}

func (u *TestUser) FlushEvents() {
	for {
		select {
		case <-u.Events:
			continue
		case <-time.After(50 * time.Millisecond): // Increase this if there is network latency
			return
		}
	}
}
