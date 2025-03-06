package main

import (
	"encoding/json"
	"log/slog"
	"strconv"
	"time"
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
		slog.Error("No board found in RegisterEvent broadcast", "board", i.Group)
		return
	}
	cols, colsOk := h.redis.GetBoardColumns(i.Group)
	if !colsOk {
		slog.Error("No board columns found in RegisterEvent broadcast", "board", i.Group)
		return
	}
	users, usersOk := h.redis.GetUsersPresence(i.Group)
	if !usersOk {
		slog.Error("Error when getting user presence in RegisterEvent broadcast", "board", i.Group)
		return
	}
	messages, messagesOk := h.redis.GetMessages(i.Group)
	if !messagesOk {
		slog.Error("Error getting messages in RegisterEvent broadcast", "board", i.Group)
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

	// Prepare timer details
	remainingTimeInSeconds := board.TimerExpiresAtUtc - time.Now().UTC().Unix()
	if remainingTimeInSeconds <= 0 {
		remainingTimeInSeconds = 0
	}

	// Prepare response
	response := RegisterResponse{
		Type:                  "reg",
		BoardName:             board.Name,
		BoardTeam:             board.Team,
		BoardColumns:          cols,
		BoardStatus:           board.Status.String(),
		BoardMasking:          board.Mask,
		BoardLock:             board.Lock,
		Users:                 userDetails,
		Messages:              messagesDetails,
		TimerExpiresInSeconds: uint16(remainingTimeInSeconds), // This shouldn't error out since we will restrict expiry to max 1 hour (3600 seconds) future time, when saving "board.TimerExpiresAtUtc".
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
		slog.Error("Error marshalling UserClosingEvent", "details", err.Error(), "payload", p)
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
		slog.Warn("Cannot find board when handling MaskEvent", "board", p.Group)
		return
	}
	// validate
	if b.Owner != p.By {
		slog.Warn("Non-owner trying to update board when handling MaskEvent", "board", p.Group, "user", p.By)
		return
	}
	if b.Mask == p.Mask {
		slog.Warn("Skipping. Board is already masked/unmasked")
		return
	}
	// Update
	if updated := h.redis.UpdateMasking(b, p.Mask); !updated {
		slog.Warn("Skipping. Unable to update masking information.")
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

type LockEvent struct {
	By    string `json:"by"`
	Group string `json:"grp"`
	Lock  bool   `json:"lock"`
}

func (p *LockEvent) Handle(i *Event, h *Hub) {
	// Update Redis
	b, ok := h.redis.GetBoard(p.Group)
	if !ok {
		slog.Warn("Cannot find board when handling LockEvent", "board", p.Group)
		return
	}
	// validate
	if b.Owner != p.By {
		slog.Warn("Non-owner trying to update board when handling LockEvent", "board", p.Group, "user", p.By)
		return
	}
	if b.Lock == p.Lock {
		slog.Warn("Skipping. Board is already locked/unlocked")
		return
	}
	// Update
	if updated := h.redis.UpdateBoardLock(b, p.Lock); !updated {
		slog.Warn("Skipping. Unable to update lock information.")
		return
	}
	// Publish to Redis (for broadcasting)
	// *Message is nil as this is not a message related update. Locking is a UI gimmick. Find a better way.
	h.redis.Publish(b.Id, &BroadcastArgs{Message: nil, Event: i})
}
func (i *LockEvent) Broadcast(h *Hub) {
	// Transform to Outgoing format
	// We can trust the LockEvent.Lock payload here. The Handle must have validated it. Don't want to add another field in BroadcastArgs{}.
	response := &LockResponse{Type: "lock", Lock: i.Lock}

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
	Anonymous  bool   `json:"anon"`
}

func (p *MessageEvent) Handle(i *Event, h *Hub) {
	// Validation: Do not allow updates in read-only board
	if isLocked := h.redis.IsBoardLocked(p.Group); isLocked {
		slog.Warn("Cannot save message in read-only board", "board", p.Group)
		return
	}
	// Save to Redis
	saved := false
	msg, exists := h.redis.GetMessage(p.Id)
	// If message doesn't exist in Redis, consider it new and save it.
	if !exists {
		msg = p.ToMessage()
		saved = h.redis.Save(msg)
	}
	if exists {
		// Only Content can be updated. The remaining fields must not be modified.
		// Validation: User can only update own message.
		// Validation: Board owner can update any message.
		isBoardOwner := h.redis.IsBoardOwner(p.Group, p.By) // Todo: No need to run this evey time. Check.
		if msg.Id == p.Id && msg.Group == p.Group && (msg.By == p.By || isBoardOwner) {
			msg.Content = p.Content
			// msg.Category = p.Category
			saved = h.redis.Save(msg)
		} else {
			slog.Warn("Cannot update someone else's message in MessageEvent handle", "msgId", p.Id, "user", p.By)
		}
	}
	if !saved {
		slog.Warn("Failed to save message in MessageEvent handle", "msgId", msg.Id)
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
		response.Mine = client.id == m.By                  //client.id == i.By. Since board owner can also update message. "mine" should reflect owner of the message.
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
		slog.Warn("Message doesn't exist in LikeMessageEvent handle", "msgId", p.MessageId)
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
	// Validation: Do not allow updates in read-only board
	if isLocked := h.redis.IsBoardLocked(p.Group); isLocked {
		slog.Warn("Cannot delete message in read-only board", "board", p.Group)
		return
	}
	// Update Redis
	deleted := false
	msg, exists := h.redis.GetMessage(p.MessageId)
	if !exists {
		slog.Warn("Message doesn't exist in DeleteMessageEvent handle", "msgId", p.MessageId)
		return
	}
	// Validate before deleting; especially if the message being deleted is of the user who created/owns it.
	// Board owner can delete any message.
	isBoardOwner := h.redis.IsBoardOwner(p.Group, p.By)
	if msg.Id == p.MessageId && msg.Group == p.Group && (msg.By == p.By || isBoardOwner) {
		if deleted = h.redis.DeleteMessage(msg); !deleted {
			return
		}
	} else {
		slog.Warn("Cannot delete someone else's message", "msgId", p.MessageId, "user", p.By)
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

type CategoryChangeEvent struct {
	MessageId   string `json:"msgId"`
	By          string `json:"by"`
	Group       string `json:"grp"`
	NewCategory string `json:"newcat"`
	OldCategory string `json:"oldcat"`
}

func (p *CategoryChangeEvent) Handle(i *Event, h *Hub) {
	// Validation: Do not allow updates in read-only board
	if isLocked := h.redis.IsBoardLocked(p.Group); isLocked {
		slog.Warn("Cannot change message category in read-only board", "board", p.Group)
		return
	}
	// Update Redis
	updated := false
	msg, exists := h.redis.GetMessage(p.MessageId)
	if !exists {
		slog.Warn("Message doesn't exist in CategoryChangeEvent handle", "msgId", p.MessageId)
		return
	}
	// Validate if new and old categories are different
	if msg.Category == p.NewCategory {
		slog.Warn("Old and New categories are same. Not changing.", "msgId", p.MessageId, "newCategory", p.NewCategory)
		return
	}
	// Validate before changing category; especially if the message being moved is of the user who created/owns it.
	// Board owner can change category of any message.
	isBoardOwner := h.redis.IsBoardOwner(p.Group, p.By)
	if msg.Id == p.MessageId && msg.Group == p.Group && (msg.By == p.By || isBoardOwner) {
		// msg.Category = p.NewCategory
		if updated = h.redis.UpdateMessageCategory(p.MessageId, p.NewCategory); !updated {
			return
		}
	} else {
		slog.Warn("Cannot change category of someone else's message", "msgId", p.MessageId, "user", p.By)
		return
	}
	// Publish to Redis (for broadcasting)
	// *Message is nil as all message details need not be broadcasted. Event details should be enough.
	if updated {
		h.redis.Publish(msg.Group, &BroadcastArgs{Message: nil, Event: i})
	}
}
func (i *CategoryChangeEvent) Broadcast(h *Hub) {
	// Transform to Outgoing format
	// We can trust the "i" *CategoryChangeResponse payload here. The Handle must have validated it. Don't want to add another field in BroadcastArgs{}.
	response := &CategoryChangeResponse{Type: "catchng", MessageId: i.MessageId, NewCategory: i.NewCategory}

	clients := h.clients[i.Group]
	for client := range clients {
		select {
		case client.send <- response:
		default:
			client.hub.unregister <- client
		}
	}
}

type TimerEvent struct {
	By                      string `json:"by"`
	Group                   string `json:"grp"`
	ExpiryDurationInSeconds uint16 `json:"expiryDurationInSeconds"`
	Stop                    bool   `json:"stop"`
}

func (p *TimerEvent) Handle(i *Event, h *Hub) {
	// The TimerEventHandle handles 2 things separately
	// 	"Stop" - Is used to Stop the timer. When this is true, "ExpiryDurationInSeconds" passed in payload is ignored.
	// 	"ExpiryDurationInSeconds" - Is used to convey the "Start" of a timer.
	// Both are mutually exclusive mostly

	// Update Redis
	b, ok := h.redis.GetBoard(p.Group)
	if !ok {
		slog.Warn("Cannot find board when handling TimerEvent", "board", p.Group)
		return
	}
	// Validate for both
	if b.Owner != p.By {
		slog.Warn("Non-owner trying to update board when handling TimerEvent", "board", p.Group, "user", p.By)
		return
	}

	// Validate and Execute for "Stop"
	// Avoid executing "Stop" unnecessarily.
	// Since a "Stop" updates the board's "timerExpiresAtUtc" downsteam, its better to avoid running it when "timerExpiresAtUtc" is 0.
	// "timerExpiresAtUtc" has its initial value of 0. A higher value means we can deduce that a timer has been set atleast once...
	// ...This validation prevents us from loosing that capability to deduce.
	if p.Stop && b.TimerExpiresAtUtc == 0 {
		slog.Warn("Cannot stop a Timer that hasn't yet started", "board", p.Group)
		return
	}

	// Execute Stop timer
	if p.Stop {
		if updated := h.redis.StopTimer(b); !updated {
			slog.Warn("Skipping. Unable to update timer during 'Stop' operation.")
			return
		}
	}

	// Validate and Execute for updating timer a.k.a "START"
	// Duration check validation is only when the board owner is trying to set the timer. "Stop" is ignored here.
	if !p.Stop && (p.ExpiryDurationInSeconds < 1 && p.ExpiryDurationInSeconds > 3600) {
		slog.Warn("Invalid timer duration. Valid duration range is between 1 and 3600 seconds.")
		return
	}

	// Execute Update timer
	if !p.Stop {
		if updated := h.redis.UpdateTimer(b, p.ExpiryDurationInSeconds); !updated {
			slog.Warn("Skipping. Unable to update timer information.")
			return
		}
	}

	// Publish to Redis (for broadcasting)
	// *Message is nil as this is not a message related update. Timer is a UI gimmick. Find a better way.
	h.redis.Publish(b.Id, &BroadcastArgs{Message: nil, Event: i})
}
func (i *TimerEvent) Broadcast(h *Hub) {
	// Transform to Outgoing format
	// redis.Getboard() is called twice in Handle and Broadcast. Let it be that way for now. Don't want to add another field in BroadcastArgs{}.
	board, boardOk := h.redis.GetBoard(i.Group)
	if !boardOk {
		slog.Warn("Cannot find board when broadcasting TimerEvent", "board", i.Group)
		return
	}

	// Prepare timer details
	remainingTimeInSeconds := board.TimerExpiresAtUtc - time.Now().UTC().Unix()
	if remainingTimeInSeconds <= 0 {
		remainingTimeInSeconds = 0
	}

	// uint16: This shouldn't error out since we will restrict expiry to max 1 hour (3600 seconds) future time, when saving "board.TimerExpiresAtUtc".
	response := &TimerResponse{Type: "timer", ExpiresInSeconds: uint16(remainingTimeInSeconds)}

	clients := h.clients[i.Group]
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
