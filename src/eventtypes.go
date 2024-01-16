package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

type RegisterEvent struct {
	By    string `json:"by"`
	Group string `json:"grp"`
}

func (e *RegisterEvent) Handle(c *Client) {
	// Note: This is not called from Hub. Avoid accessing data structures here that may run into concurrency issues.
	// Set the client id
	if c.id == "" && e.By != "" {
		c.id = e.By
	}
	b, ok := c.hub.redis.GetBoard(e.Group)
	if !ok {
		return
	}
	// No broadcast involved here. The response is just sent back to the client websocket which dispatched the RegisterEvent.
	if c.id == e.By {
		// Prepare response
		response := RegisterResponse{
			Type:         "reg",
			BoardName:    b.Name,
			BoardTeam:    b.Team,
			BoardStatus:  b.Status.String(),
			BoardMasking: b.Mask,
			IsBoardOwner: b.Owner == e.By,
		}
		select {
		case c.send <- response:
		default:
			c.hub.unregister <- c
		}
	}
}

type PresentEvent struct {
	By         string `json:"by"`
	ByNickname string `json:"nickname"`
	Xid        string `json:"xid"`
	Group      string `json:"grp"`
	Present    bool   `json:"present"`
}

func (p *PresentEvent) Handle(i *Event, h *Hub) {
	// Update Redis

	// validate
	if p.By == "" {
		return
	}

	ok := h.redis.CommitUserPresence(p.Group, &User{Id: p.By, Xid: p.Xid, Nickname: p.ByNickname}, p.Present)
	if !ok {
		return
	}

	// Publish to Redis (for broadcasting)
	// *Message is nil as this is not a message related update. Find a better way.
	h.redis.Publish(p.Group, &BroadcastArgs{Message: nil, Event: i})
}
func (i *PresentEvent) Broadcast(h *Hub) {
	// Transform to Outgoing format
	// Don't want to add another field in BroadcastArgs{}.

	users, ok := h.redis.GetUsersPresence(i.Group)
	if !ok {
		return
	}

	res := make([]*PresentDetails, 0)
	for _, user := range users {
		if user.Nickname == "" {
			user.Nickname = "Anonymous" // This may not happen.
		}
		res = append(res, &PresentDetails{Nickname: user.Nickname, Xid: user.Xid})
	}
	response := &PresentResponse{Type: "present", Users: res}

	clients := h.clients[i.Group]
	for client := range clients {
		select {
		case client.send <- response:
		default:
			client.hub.unregister <- client
		}
	}
}

// Todo: RemoveUserPresenceEvent is not a clean implementation. Refactor.
// Not directly initiated from UI. This is used during connection close. Usually when the user closes the tab or browser window.
// This is used to return a PresentResponse.
type RemoveUserPresenceEvent struct {
	By    string `json:"by"`
	Group string `json:"grp"`
}

// RemoveUserPresenceEvent is initiated from the clients Read() goroutine when its closing. Not from UI
func (p *RemoveUserPresenceEvent) Handle(h *Hub) {
	// Update Redis
	// Validate
	if p.By == "" {
		return
	}
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
	var ev = &Event{Type: "present", Payload: json.RawMessage(jsonifiedEvent)}
	// Bad hack end

	h.redis.Publish(p.Group, &BroadcastArgs{Message: nil, Event: ev})
}
func (i *RemoveUserPresenceEvent) Broadcast(h *Hub) {
	// Returning a PresentResponse. There won't be a "RemoveUserPresenceResponse".
	// Duplicate of PresentEvent Broadcast.
	users, ok := h.redis.GetUsersPresence(i.Group)
	if !ok {
		return
	}

	res := make([]*PresentDetails, 0)
	for _, user := range users {
		if user.Nickname == "" {
			user.Nickname = "Anonymous" // This may not happen.
		}
		res = append(res, &PresentDetails{Nickname: user.Nickname, Xid: user.Xid})
	}
	response := &PresentResponse{Type: "present", Users: res}

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
