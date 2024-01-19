package main

type UserDetails struct {
	Nickname string `json:"nickname"`
	Xid      string `json:"xid"`
}

type RegisterResponse struct {
	Type         string         `json:"typ"`
	Mine         bool           `json:"mine"`
	BoardName    string         `json:"boardName"`
	BoardTeam    string         `json:"boardTeam"`
	BoardStatus  string         `json:"boardStatus"`
	BoardMasking bool           `json:"boardMasking"`
	IsBoardOwner bool           `json:"isBoardOwner"`
	Users        []*UserDetails `json:"users"`
}

type UserClosingResponse struct {
	Type  string         `json:"typ"`
	Users []*UserDetails `json:"users"`
}

type MaskResponse struct {
	Type string `json:"typ"`
	Mask bool   `json:"mask"`
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
