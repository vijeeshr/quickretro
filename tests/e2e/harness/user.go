package harness

import (
	"crypto/tls"
	"encoding/json"
	"errors"
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
	Xid      string
	Nickname string
	Board    string
	Conn     *websocket.Conn
	Received chan Event
	Done     chan struct{}
}

func NewUser(id, nickname, board string) *TestUser {
	return &TestUser{
		Id:       id,
		Xid:      "xid-" + id,
		Nickname: nickname,
		Board:    board,
		Received: make(chan Event, 100),
		Done:     make(chan struct{}),
	}
}

// func (u *TestUser) NextMessageID() string {
// 	return fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), u.Id)
// }

// func (u *TestUser) NextCommentID() string {
// 	return fmt.Sprintf("cmt-%d-%s", time.Now().UnixNano(), u.Id)
// }

func (u *TestUser) Connect(baseUrl string) error {
	url := fmt.Sprintf("%s/ws/board/%s/user/%s/meet?xid=%s", baseUrl, u.Board, u.Id, u.Xid)
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
			u.Received <- Event{
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

func (u *TestUser) SendEvent(typ string, payload any) error {
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
		ByNickname: u.Nickname,
	}
	return u.SendEvent("reg", reg)
}

func (u *TestUser) SendMessage(id, content, category string) error {
	msg := MessageEvent{
		Id:         id,
		ByNickname: u.Nickname,
		Content:    content,
		Category:   category,
		ParentId:   "",
		Anonymous:  false,
	}
	return u.SendEvent("msg", msg)
}

func (u *TestUser) SendAnonymousMessage(id, content, category, nickname string, anonymous bool) error {
	msg := MessageEvent{
		Id:         id,
		ByNickname: nickname,
		Content:    content,
		Category:   category,
		ParentId:   "",
		Anonymous:  anonymous,
	}
	return u.SendEvent("msg", msg)
}

func (u *TestUser) SendComment(id, content, category, ParentMessageId string) error {
	msg := MessageEvent{
		Id:         id,
		ByNickname: u.Nickname,
		Content:    content,
		Category:   category,
		ParentId:   ParentMessageId,
		Anonymous:  false,
	}
	return u.SendEvent("msg", msg)
}

// Delete message with no comments
func (u *TestUser) DeleteMessage(msgId string) error {
	return u.deleteMessage(msgId, nil)
}

// Delete message and its known comments
func (u *TestUser) DeleteMessageWithComments(msgId string, commentIds ...string) error {
	return u.deleteMessage(msgId, commentIds)
}

func (u *TestUser) DeleteComment(cmtId string) error {
	return u.deleteMessage(cmtId, nil)
}

func (u *TestUser) deleteMessage(msgId string, commentIds []string) error {
	delEv := DeleteMessageEvent{
		MessageId:  msgId,
		CommentIds: commentIds,
	}
	return u.SendEvent("del", delEv)
}

func (u *TestUser) DeleteBoard() error {
	delAllEv := DeleteAllEvent{}
	return u.SendEvent("delall", delAllEv)
}

func (u *TestUser) Mask(mask bool) error {
	maskEv := MaskEvent{
		Mask: mask,
	}
	return u.SendEvent("mask", maskEv)
}

func (u *TestUser) LockBoard(lock bool) error {
	lockEv := LockEvent{
		Lock: lock,
	}
	return u.SendEvent("lock", lockEv)
}

func (u *TestUser) ChangeCategoryOfMessage(msgId, oldCategory, newCategory string) error {
	return u.changeMessageCategory(msgId, oldCategory, newCategory, nil)
}

func (u *TestUser) ChangeCategoryOfMessageAndComments(msgId, oldCategory, newCategory string, commentIds ...string) error {
	return u.changeMessageCategory(msgId, oldCategory, newCategory, commentIds)
}

func (u *TestUser) changeMessageCategory(msgId, oldCategory, newCategory string, commentIds []string) error {
	changeCatEv := CategoryChangeEvent{
		MessageId:   msgId,
		OldCategory: oldCategory,
		NewCategory: newCategory,
		CommentIds:  commentIds,
	}
	return u.SendEvent("catchng", changeCatEv)
}

func (u *TestUser) LikeMessage(msgId string, like bool) error {
	likeEv := LikeMessageEvent{
		MessageId: msgId,
		Like:      like,
	}
	return u.SendEvent("like", likeEv)
}

func (u *TestUser) StartTimer(expiryDurationInSeconds uint16) error {
	timerEv := TimerEvent{
		ExpiryDurationInSeconds: expiryDurationInSeconds,
		Stop:                    false,
	}
	return u.SendEvent("timer", timerEv)
}

func (u *TestUser) StopTimer(expiryDurationInSeconds uint16) error {
	timerEv := TimerEvent{
		ExpiryDurationInSeconds: expiryDurationInSeconds,
		Stop:                    true,
	}
	return u.SendEvent("timer", timerEv)
}

func (u *TestUser) SendTyping() error {
	return u.SendEvent("t", TypedEvent{})
}

func (u *TestUser) ChangeColumns(cols []*BoardColumn) error {
	colChangeEv := ColumnsChangeEvent{
		Columns: cols,
	}
	return u.SendEvent("colreset", colChangeEv)
}

func (u *TestUser) MustWaitForEvent(t *testing.T, eventType string, target any) {
	t.Helper()

	timeout := time.After(2 * time.Second)
	for {
		select {
		case ev := <-u.Received:
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

var ErrUnexpectedEvent = errors.New("unexpected event received")

func (u *TestUser) MustNotReceiveEvent(eventType string) error {
	timeout := time.After(500 * time.Millisecond)
	for {
		select {
		case ev := <-u.Received:
			if ev.Type == eventType {
				return fmt.Errorf("%w: type=%s payload=%s", ErrUnexpectedEvent, ev.Type, string(ev.Raw))
				// t.Fatalf("unexpected event received: %s (payload=%s)", eventType, string(ev.Raw))
			}
			// Ignore other event types
		case <-timeout:
			return nil // Success: no matching event received
		}
	}
}

func (u *TestUser) MustNotReceiveAnyEvent() error {
	timeout := time.After(500 * time.Millisecond)
	select {
	case ev := <-u.Received:
		return fmt.Errorf("%w: type=%s payload=%s", ErrUnexpectedEvent, ev.Type, string(ev.Raw))
		// t.Fatalf("unexpected event received: %s (payload=%s)", ev.Type, string(ev.Raw))
	case <-timeout:
		return nil // success
	}
}

func (u *TestUser) FlushEvents() {
	for {
		select {
		case <-u.Received:
			continue
		case <-time.After(50 * time.Millisecond): // Increase this if there is network latency
			return
		}
	}
}
