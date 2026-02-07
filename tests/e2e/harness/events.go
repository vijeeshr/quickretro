package harness

import "encoding/json"

// Event represents the WebSocket event structure
type Event struct {
	Type    string          `json:"typ"`
	Payload json.RawMessage `json:"pyl"`
	Raw     json.RawMessage `json:"-"`
}

type RegisterEvent struct {
	ByNickname string `json:"nickname"`
}
type RegisterResponse struct {
	Type                      string            `json:"typ"`
	BoardColumns              []*BoardColumn    `json:"columns"`
	Users                     []UserDetails     `json:"users"`
	Messages                  []MessageResponse `json:"messages"`
	Comments                  []MessageResponse `json:"comments"`
	BoardExpiryTimeUtcSeconds int64             `json:"boardExpiryUtcSeconds"` // Unix Timestamp Seconds
	TimerExpiresInSeconds     uint16            `json:"timerExpiresInSeconds"` // uint16 since we are restricting timer to max 1 hour (3600 seconds)
	NotifyNewBoardExpiry      bool              `json:"notifyNewBoardExpiry"`
	BoardMasking              bool              `json:"boardMasking"`
	BoardLock                 bool              `json:"boardLock"`
	IsBoardOwner              bool              `json:"isBoardOwner"`
}
type BoardColumn struct {
	Id        string `redis:"id" json:"id"`
	Text      string `redis:"text" json:"text"`
	Color     string `redis:"color" json:"color"`
	Position  int    `redis:"pos" json:"pos"`
	IsDefault bool   `redis:"isDefault" json:"isDefault"`
}
type UserDetails struct {
	Nickname string `json:"nickname"`
	Xid      string `json:"xid"`
}
type UserJoiningResponse struct {
	Type     string `json:"typ"`
	Nickname string `json:"nickname"`
	Xid      string `json:"xid"`
}
type UserClosingResponse struct {
	Type string `json:"typ"`
	Xid  string `json:"xid"`
}

type MessageEvent struct {
	Id         string `json:"id"`
	ByNickname string `json:"nickname"`
	Content    string `json:"msg"`
	Category   string `json:"cat"`
	ParentId   string `json:"pid"`
	Anonymous  bool   `json:"anon"`
}
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

type DeleteMessageEvent struct {
	MessageId  string   `json:"msgId"`      // MessageId or CommentId
	CommentIds []string `json:"commentIds"` // Only used when deleting a top-level message i.e. when MessageId represents a message and not a comment.
}
type DeleteMessageResponse struct {
	Id string `json:"id"`
}

type DeleteAllEvent struct {
}
type DeleteAllResponse struct {
	Type string `json:"typ"`
}

type MaskEvent struct {
	Mask bool `json:"mask"`
}
type MaskResponse struct {
	Type string `json:"typ"`
	Mask bool   `json:"mask"`
}

type LockEvent struct {
	Lock bool `json:"lock"`
}
type LockResponse struct {
	Type string `json:"typ"`
	Lock bool   `json:"lock"`
}

type CategoryChangeEvent struct {
	MessageId   string   `json:"msgId"`
	NewCategory string   `json:"newcat"`
	OldCategory string   `json:"oldcat"`
	CommentIds  []string `json:"commentIds"`
}
type CategoryChangeResponse struct {
	Type        string `json:"typ"`
	MessageId   string `json:"id"`
	NewCategory string `json:"newcat"`
}

type LikeMessageEvent struct {
	MessageId string `json:"msgId"`
	Like      bool   `json:"like"`
}
type LikeMessageResponse struct {
	Type  string `json:"typ"`
	Id    string `json:"id"`
	Likes int64  `json:"likes"`
	Liked bool   `json:"liked"` // True if receiving user has liked this message.
}

type TimerEvent struct {
	ExpiryDurationInSeconds uint16 `json:"expiryDurationInSeconds"`
	Stop                    bool   `json:"stop"`
}
type TimerResponse struct {
	Type             string `json:"typ"`
	ExpiresInSeconds uint16 `json:"expiresInSeconds"`
}

type ColumnsChangeEvent struct {
	Columns []*BoardColumn `json:"columns"`
	// Only columns to add/update are sent. Columns to disable aren't sent explicitly.
}
type ColumnsChangeResponse struct {
	Type         string         `json:"typ"`
	BoardColumns []*BoardColumn `json:"columns"`
}

type BroadcastArgs struct {
	Event Event `json:"event"`
	// Message *Message `json:"message,omitempty"` // simplified
}

type TypedEvent struct{}

type TypedResponse struct {
	Type string `json:"typ"`
	Xid  string `json:"xid"`
}
