package main

// Store
type Message struct {
	Id         string `redis:"id"`
	By         string `redis:"by"`
	ByNickname string `redis:"nickname"`
	Group      string `redis:"group"`
	Content    string `redis:"content"`
	Category   string `redis:"category"`
	Anonymous  bool   `redis:"anon"`
}

func (m *MessageEvent) ToMessage() *Message {
	return &Message{
		Id: m.Id, By: m.By, ByNickname: m.ByNickname, Group: m.Group, Content: m.Content, Category: m.Category, Anonymous: m.Anonymous}
}

func (m *Message) NewResponse(reqType string) interface{} {
	switch reqType {
	case "del":
		return DeleteMessageResponse{
			Type: reqType,
			Id:   m.Id,
		}
	case "like":
		return LikeMessageResponse{
			Type: reqType,
			Id:   m.Id,
		}
	default:
		return MessageResponse{
			Type:       reqType,
			Id:         m.Id,
			ByNickname: m.ByNickname,
			Content:    m.Content,
			Category:   m.Category,
			Anonymous:  m.Anonymous,
		}
	}
}
