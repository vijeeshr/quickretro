package main

type RegisterResponse struct {
	Type         string `json:"typ"`
	BoardName    string `json:"boardName"`
	BoardTeam    string `json:"boardTeam"`
	BoardStatus  string `json:"boardStatus"`
	BoardMasking bool   `json:"boardMasking"`
	IsBoardOwner bool   `json:"isBoardOwner"`
}

type MaskResponse struct {
	Type string `json:"typ"`
	Mask bool   `json:"mask"`
}

type PresentResponse struct {
	Type  string            `json:"typ"`
	Users []*PresentDetails `json:"users"`
}
type PresentDetails struct {
	Nickname string `json:"nickname"`
	Xid      string `json:"xid"`
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
