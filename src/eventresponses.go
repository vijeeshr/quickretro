package main

type UserDetails struct {
	Nickname string `json:"nickname"`
	Xid      string `json:"xid"`
}

type RegisterResponse struct {
	Type                      string            `json:"typ"`
	Mine                      bool              `json:"mine"`
	BoardName                 string            `json:"boardName"`
	BoardTeam                 string            `json:"boardTeam"`
	BoardColumns              []*BoardColumn    `json:"columns"` // Using same BoardColumn struct that is used for request and redis store. Todo - refactor later.
	BoardStatus               string            `json:"boardStatus"`
	BoardMasking              bool              `json:"boardMasking"`
	BoardLock                 bool              `json:"boardLock"`
	IsBoardOwner              bool              `json:"isBoardOwner"`
	Users                     []*UserDetails    `json:"users"`
	Messages                  []MessageResponse `json:"messages"`              //Todo: Change to *MessageResponse
	Comments                  []MessageResponse `json:"comments"`              //Todo: Change to *MessageResponse
	TimerExpiresInSeconds     uint16            `json:"timerExpiresInSeconds"` // uint16 since we are restricting timer to max 1 hour (3600 seconds)
	BoardExpiryTimeUtcSeconds int64             `json:"boardExpiryUtcSeconds"` // Unix Timestamp Seconds
	NotifyNewBoardExpiry      bool              `json:"notifyNewBoardExpiry"`
}

type UserClosingResponse struct {
	Type  string         `json:"typ"`
	Users []*UserDetails `json:"users"`
}

type MaskResponse struct {
	Type string `json:"typ"`
	Mask bool   `json:"mask"`
}

type LockResponse struct {
	Type string `json:"typ"`
	Lock bool   `json:"lock"`
}

type MessageResponse struct {
	Type       string `json:"typ"`
	Id         string `json:"id"`
	ByNickname string `json:"nickname"`
	Content    string `json:"msg"`
	Category   string `json:"cat"`
	Likes      string `json:"likes"`
	Liked      bool   `json:"liked"` // True if receiving user has liked this message.
	Mine       bool   `json:"mine"`
	Anonymous  bool   `json:"anon"`
	ParentId   string `json:"pid"`
}

type LikeMessageResponse struct {
	Type  string `json:"typ"`
	Id    string `json:"id"`
	Likes string `json:"likes"`
	Liked bool   `json:"liked"` // True if receiving user has liked this message.
}

type DeleteMessageResponse struct {
	Type string `json:"typ"`
	Id   string `json:"id"`
}

type DeleteAllResponse struct {
	Type string `json:"typ"`
}

type CategoryChangeResponse struct {
	Type        string `json:"typ"`
	MessageId   string `json:"id"`
	NewCategory string `json:"newcat"`
}

type TimerResponse struct {
	Type             string `json:"typ"`
	ExpiresInSeconds uint16 `json:"expiresInSeconds"`
}
