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
	ParentId   string `redis:"pid"` // For top-level "Message" this will be empty. For a message treated as "Comment", it will be the parent MessageId.
}

func (m *MessageEvent) ToMessage() *Message {
	return &Message{
		Id: m.Id, By: m.By, ByNickname: m.ByNickname, Group: m.Group, Content: m.Content, Category: m.Category, Anonymous: m.Anonymous, ParentId: m.ParentId}
}

func (m *Message) NewMessageResponse() MessageResponse {
	return MessageResponse{
		Type:       "msg",
		Id:         m.Id,
		ByNickname: m.ByNickname,
		Content:    m.Content,
		Category:   m.Category,
		Anonymous:  m.Anonymous,
		ParentId:   m.ParentId,
	}
}
func (m *Message) NewDeleteResponse() DeleteMessageResponse {
	return DeleteMessageResponse{
		Type: "del",
		Id:   m.Id,
	}
}
func (m *Message) NewLikeResponse() LikeMessageResponse {
	return LikeMessageResponse{
		Type: "like",
		Id:   m.Id,
	}
}

// Enum SaveMode
type SaveMode int

const (
	AsNewMessage SaveMode = iota
	AsNewComment
)

// func (m *Message) NewResponse(reqType string) interface{} {
// 	switch reqType {
// 	case "del":
// 		return DeleteMessageResponse{
// 			Type: reqType,
// 			Id:   m.Id,
// 		}
// 	case "like":
// 		return LikeMessageResponse{
// 			Type: reqType,
// 			Id:   m.Id,
// 		}
// 	default:
// 		return MessageResponse{
// 			Type:       reqType,
// 			Id:         m.Id,
// 			ByNickname: m.ByNickname,
// 			Content:    m.Content,
// 			Category:   m.Category,
// 			Anonymous:  m.Anonymous,
// 			ParentId:   m.ParentId,
// 		}
// 	}
// }
