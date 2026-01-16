package harness

import "encoding/json"

// Event represents the WebSocket event structure
type Event struct {
	Type    string          `json:"typ"`
	Payload json.RawMessage `json:"pyl"`
	Raw     json.RawMessage `json:"-"`
}

// RegisterEvent payload
type RegisterEvent struct {
	By         string `json:"by"`
	ByNickname string `json:"nickname"`
	Xid        string `json:"xid"`
	Group      string `json:"grp"`
}

// RegisterResponse payload
type RegisterResponse struct {
	Type         string `json:"typ"`
	IsBoardOwner bool   `json:"isBoardOwner"`
	// Add other fields if needed for verification
}

type UserJoiningResponse struct {
	Type     string `json:"typ"`
	Nickname string `json:"nickname"`
	Xid      string `json:"xid"`
}

// MessageEvent payload
type MessageEvent struct {
	Id         string `json:"id"`
	By         string `json:"by"`
	ByXid      string `json:"byxid"`
	ByNickname string `json:"nickname"`
	Group      string `json:"grp"`
	Content    string `json:"msg"`
	Category   string `json:"cat"`
	ParentId   string `json:"pid"`
	Anonymous  bool   `json:"anon"`
}

// MessageResponse payload (for receiving)
type MessageResponse struct {
	Type       string `json:"typ"`
	Id         string `json:"id"`
	ParentId   string `json:"pid"`
	ByXid      string `json:"byxid"`
	ByNickname string `json:"nickname"`
	Content    string `json:"msg"`
	Category   string `json:"cat"`
	Likes      int64  `json:"likes"`
	Liked      bool   `json:"liked"` // True if receiving user has liked this message.
	Mine       bool   `json:"mine"`
	Anonymous  bool   `json:"anon"`
}

// LikeMessageEvent payload
type LikeMessageEvent struct {
	MessageId string `json:"msgId"`
	By        string `json:"by"`
	Like      bool   `json:"like"`
}

// LockEvent payload
type LockEvent struct {
	By    string `json:"by"`
	Group string `json:"grp"`
	Lock  bool   `json:"lock"`
}

type BroadcastArgs struct {
	Event Event `json:"event"`
	// Message *Message `json:"message,omitempty"` // simplified
}
