package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

type RegisterEvent struct {
	By         string `json:"by"`
	ByNickname string `json:"nickname"`
	Xid        string `json:"xid"`
	Group      string `json:"grp"`
}

func (p *RegisterEvent) Handle(i *Event, h *Hub) {
	// Validate
	// We can be sure that the board exists. That check is done during handshake.
	if p.By == "" || p.Xid == "" || p.Group == "" {
		return
	}
	// Update Redis
	if ok := h.redis.CommitUserPresence(p.Group, &User{Id: p.By, Xid: p.Xid, Nickname: p.ByNickname}, true); !ok {
		return
	}

	// Publish to Redis (for broadcasting)
	// *Message is nil as this is not a message related update. Find a better way. Generics?
	h.redis.Publish(p.Group, &BroadcastArgs{Message: nil, Event: i})
}
func (i *RegisterEvent) Broadcast(h *Hub) {
	// Transform to Outgoing format
	// Don't want to add another field in BroadcastArgs{}.
	board, boardOk := h.redis.GetBoard(i.Group)
	if !boardOk {
		return
	}
	cols, colsOk := h.redis.GetBoardColumns(i.Group)
	if !colsOk {
		return
	}
	users, usersOk := h.redis.GetUsersPresence(i.Group)
	if !usersOk {
		return
	}
	messages, messagesOk := h.redis.GetMessages(i.Group)
	if !messagesOk {
		return
	}

	// Prepare user details
	userDetails := make([]*UserDetails, 0)
	for _, user := range users {
		if user.Nickname == "" {
			user.Nickname = "Anonymous" // This may not happen.
		}
		userDetails = append(userDetails, &UserDetails{Nickname: user.Nickname, Xid: user.Xid})
	}

	// Prepare message details
	messagesDetails := make([]MessageResponse, 0)
	// Collect "like" count for all messages in one call via a Redis pipeline
	ids := make([]string, 0)
	for _, m := range messages {
		ids = append(ids, m.Id)
	}
	likes := h.redis.GetLikesCountMultiple(ids...)
	for _, m := range messages {
		msgRes := m.NewResponse("msg").(MessageResponse) // Todo: Type to return *MessageResponse
		if count, ok := likes[m.Id]; ok {
			msgRes.Likes = strconv.FormatInt(count, 10)
		}
		msgRes.Mine = m.By == i.By
		msgRes.Liked = h.redis.HasLiked(m.Id, i.By) // Todo: This calls Redis SISMEMBER [O(1) as per doc] in a loop. Check for impact.
		messagesDetails = append(messagesDetails, msgRes)
	}

	// Prepare response
	response := RegisterResponse{
		Type:         "reg",
		BoardName:    board.Name,
		BoardTeam:    board.Team,
		BoardColumns: cols,
		BoardStatus:  board.Status.String(),
		BoardMasking: board.Mask,
		Users:        userDetails,
		Messages:     messagesDetails,
	}

	clients := h.clients[i.Group]
	for client := range clients {
		response.Mine = client.id == i.By // This is to identify if RegisterResponse is a result of a RegisterEvent initiated by same user. RegisterResponses are broadcasted to all.
		response.IsBoardOwner = client.id == board.Owner
		select {
		case client.send <- response:
		default:
			client.hub.unregister <- client
		}
	}
}

// Todo: UserClosingEvent is not a clean implementation. Refactor.
// Not directly initiated from UI. This is used during connection close. Usually when the user closes the tab or browser window.
type UserClosingEvent struct {
	By    string `json:"by"`
	Group string `json:"grp"`
}

// UserClosingEvent is initiated from the clients Read() goroutine when its closing. Not from UI
func (p *UserClosingEvent) Handle(h *Hub) {
	// Validate
	if p.By == "" || p.Group == "" {
		return
	}
	// Update Redis
	if ok := h.redis.RemoveUserPresence(p.Group, p.By); !ok {
		return
	}

	// Publish to Redis (for broadcasting)

	// Bad hack start -
	jsonifiedEvent, err := json.Marshal(p)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}
	var ev = &Event{Type: "closing", Payload: json.RawMessage(jsonifiedEvent)}
	// Bad hack end

	h.redis.Publish(p.Group, &BroadcastArgs{Message: nil, Event: ev})
}
func (i *UserClosingEvent) Broadcast(h *Hub) {
	// Duplication with UserClosingEvent Broadcast.
	users, ok := h.redis.GetUsersPresence(i.Group)
	if !ok {
		return
	}

	res := make([]*UserDetails, 0)
	for _, user := range users {
		if user.Nickname == "" {
			user.Nickname = "Anonymous" // This may not happen.
		}
		res = append(res, &UserDetails{Nickname: user.Nickname, Xid: user.Xid})
	}
	response := &UserClosingResponse{Type: "closing", Users: res}

	clients := h.clients[i.Group]
	for client := range clients {
		// skip sending to the client that is closing
		if client.id != i.By {
			select {
			case client.send <- response:
			default:
				client.hub.unregister <- client
			}
		}
	}
}

type MaskEvent struct {
	By    string `json:"by"`
	Group string `json:"grp"`
	Mask  bool   `json:"mask"`
}

func (p *MaskEvent) Handle(i *Event, h *Hub) {
	// Update Redis
	b, ok := h.redis.GetBoard(p.Group)
	if !ok {
		log.Println("cannot find board")
		return
	}
	// validate
	if b.Owner != p.By {
		log.Println("only owner can update board")
		return
	}
	if b.Mask == p.Mask {
		log.Println("already masked/unmasked. skipping.")
		return
	}
	// Update
	if updated := h.redis.UpdateMasking(b, p.Mask); !updated {
		log.Println("unable to update masking information in Redis. skipping")
		return
	}
	// Publish to Redis (for broadcasting)
	// *Message is nil as this is not a message related update. Masking is a UI gimmick. Find a better way.
	h.redis.Publish(b.Id, &BroadcastArgs{Message: nil, Event: i})
}
func (i *MaskEvent) Broadcast(h *Hub) {
	// Transform to Outgoing format
	// We can trust the MaskEvent.Mask payload here. The Handle must have validated it. Don't want to add another field in BroadcastArgs{}.
	response := &MaskResponse{Type: "mask", Mask: i.Mask}

	clients := h.clients[i.Group]
	for client := range clients {
		select {
		case client.send <- response:
		default:
			client.hub.unregister <- client
		}
	}
}

type MessageEvent struct {
	Id         string `json:"id"`
	By         string `json:"by"`
	ByNickname string `json:"nickname"`
	Group      string `json:"grp"`
	Content    string `json:"msg"`
	Category   string `json:"cat"`
}

func (p *MessageEvent) Handle(i *Event, h *Hub) {
	// Save to Redis
	saved := false
	msg, exists := h.redis.GetMessage(p.Id)
	// If message doesn't exist in Redis, consider it new and save it.
	if !exists {
		msg = p.ToMessage()
		saved = h.redis.Save(msg)
	}
	if exists {
		// Only Content and Category can be updated. The remaining fields must not be modified.
		// Validation: User can only update own message.
		if msg.Id == p.Id && msg.By == p.By && msg.Group == p.Group {
			msg.Content = p.Content
			msg.Category = p.Category
			saved = h.redis.Save(msg)
		} else {
			log.Println("cannot update someone else's message")
		}
	}
	if !saved {
		log.Printf("couldn't save message with id %v", msg.Id)
		return
	}

	// Publish to Redis (for broadcasting)
	if saved {
		h.redis.Publish(msg.Group, &BroadcastArgs{Message: msg, Event: i})
	}
}
func (i *MessageEvent) Broadcast(m *Message, h *Hub) {
	// Transform to Outgoing format
	response := m.NewResponse("msg").(MessageResponse)
	count := h.redis.GetLikesCount(m.Id)
	response.Likes = strconv.FormatInt(count, 10)

	clients := h.clients[m.Group]
	for client := range clients {
		response.Mine = client.id == i.By
		response.Liked = h.redis.HasLiked(m.Id, client.id) // Todo: This calls Redis SISMEMBER [O(1) as per doc] in a loop. Check for impact.
		select {
		case client.send <- response:
		default:
			client.hub.unregister <- client
		}
	}
}

type LikeMessageEvent struct {
	MessageId string `json:"msgId"`
	By        string `json:"by"`
	Like      bool   `json:"like"`
}

func (p *LikeMessageEvent) Handle(i *Event, h *Hub) {
	// Save to Redis
	msg, exists := h.redis.GetMessage(p.MessageId) // Todo: Check if fetching a message is needed for a like. Can avoid extra calls. Also BroadcastArgs.Message may not be needed here if removed.
	if !exists {
		log.Println("message to like doesn't exist")
		return
	}
	liked := h.redis.Like(p.MessageId, p.By, p.Like)
	if !liked {
		return
	}
	// Publish to Redis (for broadcasting)
	if liked {
		h.redis.Publish(msg.Group, &BroadcastArgs{Message: msg, Event: i})
	}
}
func (i *LikeMessageEvent) Broadcast(m *Message, h *Hub) {
	// Transform to Outgoing format
	response := m.NewResponse("like").(LikeMessageResponse)
	count := h.redis.GetLikesCount(m.Id)
	response.Likes = strconv.FormatInt(count, 10)

	clients := h.clients[m.Group]
	for client := range clients {
		response.Liked = h.redis.HasLiked(m.Id, client.id) // Todo: This calls Redis SISMEMBER [O(1) as per doc] in a loop. Check for impact.
		select {
		case client.send <- response:
		default:
			client.hub.unregister <- client
		}
	}
}

type DeleteMessageEvent struct {
	MessageId string `json:"msgId"`
	By        string `json:"by"`
	Group     string `json:"grp"`
}

func (p *DeleteMessageEvent) Handle(i *Event, h *Hub) {
	// Update Redis
	deleted := false
	msg, exists := h.redis.GetMessage(p.MessageId)
	if !exists {
		log.Println("message to delete doesn't exist")
		return
	}
	// Validate before deleting; especially if the message being deleted is of the user who created/owns it.
	if msg.Id == p.MessageId && msg.By == p.By && msg.Group == p.Group {
		if deleted = h.redis.DeleteMessage(msg); !deleted {
			return
		}
	} else {
		log.Println("Cannot delete someone else's message")
		return
	}
	// Publish to Redis (for broadcasting)
	if deleted {
		h.redis.Publish(msg.Group, &BroadcastArgs{Message: msg, Event: i}) // Todo: Similar to "Like", BroadcastArgs.Message may not be needed here.
	}
}
func (i *DeleteMessageEvent) Broadcast(m *Message, h *Hub) {
	// Transform to Outgoing format
	response := m.NewResponse("del").(DeleteMessageResponse)

	clients := h.clients[m.Group]
	for client := range clients {
		select {
		case client.send <- response:
		default:
			client.hub.unregister <- client
		}
	}
}

// Helper struct from Broadcasting
type BroadcastArgs struct {
	Event   *Event
	Message *Message
}
