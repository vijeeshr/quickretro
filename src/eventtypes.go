package main

import (
	"encoding/json"
	"log/slog"
	"time"
	"unicode/utf8"
)

type RegisterEvent struct {
	// The broadcaster of RegisterEvent i.e. "Broadcast" method below sends two types of responses -
	// RegisterResponse: Sent to the client who triggered the RegisterEvent, and
	// UserJoiningResponse: Sent to all the other active clients in the board
	ByNickname string `json:"nickname"`
}

func (p *RegisterEvent) Handle(e *Event, h *Hub) {
	// Validate
	if p.ByNickname == "" {
		return
	}
	// Mostly the board should exist. That check is done during handshake.
	// This is to prevent adding UserPresence data to redis on receipt of a reg event payload with a non-existent board.
	// TODO: Check if this is needed?
	if !h.redis.BoardExists(e.Group) {
		return
	}

	// Execute
	if ok := h.redis.CommitUserPresence(e.Group, &User{Id: e.By, Xid: e.Xid, Nickname: p.ByNickname}); !ok {
		return
	}

	// Publish to Redis (for broadcasting)
	// *Message is nil as this is not a message related update. Find a better way. Generics?
	h.redis.Publish(e.Group, &BroadcastArgs{Message: nil, Event: e})
}
func (p *RegisterEvent) Broadcast(e *Event, m *Message, h *Hub) {
	data, ok := h.redis.GetBoardAggregatedData(e.Group)
	if !ok {
		slog.Error("Failed to get board aggregated data", "board", e.Group)
		return
	}

	board, cols, users, messages, comments := data.Board, data.Columns, data.Users, data.Messages, data.Comments

	// Prepare user details
	// userDetails := make([]UserDetails, 0)
	// userDetails := make([]UserDetails, 0, len(users))
	userDetails := make([]UserDetails, len(users)) // Preallocate length instead of capacity. len == cap == len(users), so can index directly.
	for in, u := range users {
		userDetails[in] = UserDetails{Nickname: u.Nickname, Xid: u.Xid}
	}

	// Prepare message details
	messagesDetails := make([]MessageResponse, len(messages))
	// Collect "likes" info
	ids := make([]string, len(messages))
	for in, m := range messages {
		ids[in] = m.Id
	}
	likesInfo, likesOk := h.redis.GetLikesInfo(e.By, ids...)
	if !likesOk {
		slog.Warn("Failed to fetch likes info")
	}
	for in, m := range messages {
		msgRes := m.NewMessageResponse()
		msgRes.Mine = m.By == e.By
		if likesOk {
			if info, ok := likesInfo[m.Id]; ok {
				msgRes.Likes = info.Count
				msgRes.Liked = info.Liked
			}
		}
		messagesDetails[in] = msgRes
	}

	// Prepare comment details
	commentDetails := make([]MessageResponse, len(comments))
	for in, c := range comments {
		cmtRes := c.NewMessageResponse()
		cmtRes.Mine = c.By == e.By
		commentDetails[in] = cmtRes
	}

	// Prepare timer details
	nowUnix := time.Now().UTC().Unix()
	remainingTimeInSeconds := int64(0)
	if board.TimerExpiresAtUtc > nowUnix {
		remainingTimeInSeconds = board.TimerExpiresAtUtc - nowUnix
	}

	// Prepare RegisterResponse
	// RegisterResponse is only sent to client who is the initiator of RegEvent
	regResponse := RegisterResponse{
		Type:                      "reg",
		BoardName:                 board.Name,
		BoardTeam:                 board.Team,
		BoardColumns:              cols,
		BoardStatus:               board.Status.String(),
		BoardMasking:              board.Mask,
		BoardLock:                 board.Lock,
		Users:                     userDetails,
		Messages:                  messagesDetails,
		Comments:                  commentDetails,
		TimerExpiresInSeconds:     uint16(remainingTimeInSeconds), // This shouldn't error out since we will restrict expiry to max 1 hour (3600 seconds) future time, when saving "board.TimerExpiresAtUtc".
		BoardExpiryTimeUtcSeconds: board.AutoDeleteAtUtc,
		NotifyNewBoardExpiry:      (nowUnix - board.CreatedAtUtc) < 10, // Prepare board expiry notification prompt for New board (less than 10 seconds)
	}
	// Prepare UserJoiningResponse
	// UserJoiningResponse is sent to all other active clients (except initiator)
	joinResp := UserJoiningResponse{
		Type:     "joining",
		Nickname: p.ByNickname,
		Xid:      e.Xid,
	}

	clients := h.clients[e.Group]

	for client := range clients {
		// // Shallow copy.
		// // Copies the struct value. The fields/slices inside ([]MessageResponse, []UserDetails, []*BoardColumn) are NOT cloned deeply, they still point to the same underlying arrays.
		// // Ensure those fields/slices aren't mutated after being sent from here.
		// response := regResponse
		if client.id == e.By {
			regResponse.IsBoardOwner = client.id == board.Owner
			select {
			case client.send <- regResponse:
			default:
				client.hub.unregister <- client
			}
			continue
		}

		select {
		case client.send <- joinResp:
		default:
			client.hub.unregister <- client
		}
	}

	// for client := range clients {
	// 	var payload any

	// 	if client.id == i.By {
	// 		payload = regResponse
	// 	} else {
	// 		payload = joinResp
	// 	}

	// 	select {
	// 	case client.send <- payload:
	// 	default:
	// 		client.hub.unregister <- client
	// 	}
	// }
}

// Todo: UserClosingEvent is not a clean implementation. Refactor.
// Not directly initiated from UI. This is used during connection close. Usually when the user closes the tab or browser window.
type UserClosingEvent struct {
	By    string `json:"by"`
	Group string `json:"grp"`
	Xid   string `json:"xid"`
}

// UserClosingEvent is initiated from the clients Read() goroutine when its closing. Not from UI
func (p *UserClosingEvent) Handle(_ *Event, h *Hub) {
	// Validate
	if p.By == "" || p.Group == "" {
		return
	}

	// Execute
	removed := h.redis.RemoveUserPresence(p.Group, p.By)
	if !removed {
		return
	}

	// Publish to Redis (for broadcasting)

	// Bad hack start -
	jsonifiedEvent, err := json.Marshal(p)
	if err != nil {
		slog.Error("Error marshalling UserClosingEvent", "err", err, "payload", p)
		return
	}
	var ev = &Event{Type: "closing", Payload: json.RawMessage(jsonifiedEvent)}
	// Bad hack end

	h.redis.Publish(p.Group, &BroadcastArgs{Message: nil, Event: ev})
}
func (p *UserClosingEvent) Broadcast(_ *Event, m *Message, h *Hub) {
	response := &UserClosingResponse{Type: "closing", Xid: p.Xid}

	clients := h.clients[p.Group]
	for client := range clients {
		// skip sending to the client that is closing
		if client.id != p.By {
			select {
			case client.send <- response:
			default:
				client.hub.unregister <- client
			}
		}
	}
}

type MaskEvent struct {
	Mask bool `json:"mask"`
}

func (p *MaskEvent) Handle(e *Event, h *Hub) {
	// Validate
	b, ok := h.redis.GetBoard(e.Group)
	if !ok {
		slog.Warn("Cannot find board when handling MaskEvent", "board", e.Group)
		return
	}
	if b.Owner != e.By {
		slog.Warn("Non-owner trying to update board when handling MaskEvent", "board", e.Group, "user", e.By)
		return
	}
	if b.Mask == p.Mask {
		slog.Warn("Skipping. Board is already masked/unmasked")
		return
	}
	// Execute
	if updated := h.redis.UpdateMasking(b, p.Mask); !updated {
		slog.Warn("Skipping. Unable to update masking information.")
		return
	}
	// Publish to Redis (for broadcasting)
	// *Message is nil as this is not a message related update. Masking is a UI gimmick. Find a better way.
	h.redis.Publish(b.Id, &BroadcastArgs{Message: nil, Event: e})
}
func (p *MaskEvent) Broadcast(e *Event, m *Message, h *Hub) {
	// Transform to Outgoing format
	// We can trust the MaskEvent.Mask payload here. The Handle must have validated it. Don't want to add another field in BroadcastArgs{}.
	response := &MaskResponse{Type: "mask", Mask: p.Mask}

	clients := h.clients[e.Group]
	for client := range clients {
		select {
		case client.send <- response:
		default:
			client.hub.unregister <- client
		}
	}
}

type LockEvent struct {
	Lock bool `json:"lock"`
}

func (p *LockEvent) Handle(e *Event, h *Hub) {
	// validate
	b, ok := h.redis.GetBoard(e.Group)
	if !ok {
		slog.Warn("Cannot find board when handling LockEvent", "board", e.Group)
		return
	}
	if b.Owner != e.By {
		slog.Warn("Non-owner trying to update board when handling LockEvent", "board", e.Group, "user", e.By)
		return
	}
	if b.Lock == p.Lock {
		slog.Warn("Skipping. Board is already locked/unlocked")
		return
	}
	// Execute
	if updated := h.redis.UpdateBoardLock(b, p.Lock); !updated {
		slog.Warn("Skipping. Unable to update lock information.")
		return
	}
	// Publish to Redis (for broadcasting)
	// *Message is nil as this is not a message related update. Locking is a UI gimmick. Find a better way.
	h.redis.Publish(b.Id, &BroadcastArgs{Message: nil, Event: e})
}
func (p *LockEvent) Broadcast(e *Event, m *Message, h *Hub) {
	// Transform to Outgoing format
	// We can trust the LockEvent.Lock payload here. The Handle must have validated it. Don't want to add another field in BroadcastArgs{}.
	response := &LockResponse{Type: "lock", Lock: p.Lock}

	clients := h.clients[e.Group]
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
	ByNickname string `json:"nickname"`
	Content    string `json:"msg"`
	Category   string `json:"cat"`
	ParentId   string `json:"pid"`
	Anonymous  bool   `json:"anon"`
}

func (p *MessageEvent) Handle(e *Event, h *Hub) {
	// Validate
	if h.redis.IsBoardLocked(e.Group) {
		slog.Warn("Cannot save message in read-only board", "board", e.Group)
		return
	}

	msg := p.ToMessage(e.By, e.Xid, e.Group)

	existing, exists := h.redis.GetMessage(msg.Id)
	saved := false

	if !exists {
		saved = handleNewMessageOrComment(msg, h)
	} else {
		saved = handleUpdate(existing, msg, h)
	}

	if !saved {
		slog.Warn("Failed to save message/comment", "msgId", msg.Id)
		return
	}

	// Publish to Redis (for broadcasting)
	if saved {
		// Other fields in the "p" payload may be tampered or different for an existing message.
		// For an existing message that is saved/updated again, just update existing.Content from payload and send it for broadcasting
		if exists {
			existing.Content = msg.Content
			msg = existing
		}
		h.redis.Publish(msg.Group, &BroadcastArgs{Message: msg, Event: e})
	}
}
func handleNewMessageOrComment(msg *Message, h *Hub) bool {
	// New message
	if isMessage(msg) {
		if msg.Anonymous {
			msg.ByXid, msg.ByNickname = "", ""
		}
		return h.redis.Save(msg, AsNewMessage)
	}

	// New Comment
	// Validate parent message exists
	parent, exists := h.redis.GetMessage(msg.ParentId)
	if !exists {
		slog.Warn("Parent message not found", "commentId", msg.Id, "parentId", msg.ParentId)
		return false
	}
	// Ensure parent is not itself a comment (prevent nesting)
	if parent.ParentId != "" {
		slog.Warn("Cannot attach a comment to another comment", "commentId", msg.Id, "parentId", msg.ParentId)
		return false
	}
	// Ensure category is same as that of parent
	msg.Category = parent.Category

	return h.redis.Save(msg, AsNewComment)
}
func handleUpdate(existing, updated *Message, h *Hub) bool {
	// Validation: user can only update own message, and only content is editable
	if existing.Id != updated.Id || existing.Group != updated.Group || existing.By != updated.By || existing.ParentId != updated.ParentId {
		slog.Warn("Unauthorized update attempt", "msgId", updated.Id, "user", updated.By)
		return false
	}

	existing.Content = updated.Content
	return h.redis.Save(existing)
}

func (p *MessageEvent) Broadcast(e *Event, m *Message, h *Hub) {
	// Transform to Outgoing format (static per broadcast)
	base := m.NewMessageResponse()
	base.Likes = h.redis.GetLikesCount(m.Id)
	// Snapshot clients for this group
	clients := h.clients[m.Group]
	clientCount := len(clients)
	// Collect all clientIds and clients
	ids := make([]string, 0, clientCount)
	clientList := make([]*Client, 0, clientCount)
	for c := range clients {
		ids = append(ids, c.id)
		clientList = append(clientList, c)
	}

	// Bulk Redis query: SMISMEMBER
	// Order of returned results corresponds to order of "ids" passed
	likedList := h.redis.HasLiked(m.Id, ids)
	for i, client := range clientList {
		// Copy the base response
		res := base
		res.Mine = client.id == m.By
		res.Liked = likedList[i]

		select {
		case client.send <- res: // Todo: check implications of sending &res to channel and benchmark
		default:
			client.hub.unregister <- client
		}
	}
}

type LikeMessageEvent struct {
	MessageId string `json:"msgId"`
	Like      bool   `json:"like"`
}

func (p *LikeMessageEvent) Handle(e *Event, h *Hub) {
	// Validate
	msg, exists := h.redis.GetMessage(p.MessageId) // Todo: Check if fetching a message is needed for a like. Can avoid extra calls. Also BroadcastArgs.Message may not be needed here if removed.
	if !exists {
		slog.Warn("Message doesn't exist in LikeMessageEvent handle", "msgId", p.MessageId)
		return
	}
	// Execute
	liked := h.redis.Like(p.MessageId, e.By, p.Like)
	if !liked {
		return
	}
	// Publish to Redis (for broadcasting)
	if liked {
		h.redis.Publish(msg.Group, &BroadcastArgs{Message: msg, Event: e})
	}
}
func (p *LikeMessageEvent) Broadcast(e *Event, m *Message, h *Hub) {
	base := m.NewLikeResponse()
	base.Likes = h.redis.GetLikesCount(m.Id)
	// Snapshot clients for this group
	clients := h.clients[m.Group]
	clientCount := len(clients)
	// Collect all clientIds and clients
	ids := make([]string, 0, clientCount)
	clientList := make([]*Client, 0, clientCount)
	for c := range clients {
		ids = append(ids, c.id)
		clientList = append(clientList, c)
	}

	// Bulk Redis query: SMISMEMBER
	// Order of returned results corresponds to order of "ids" passed
	likedList := h.redis.HasLiked(m.Id, ids)
	for i, client := range clientList {
		// Copy the base response
		resp := base
		resp.Liked = likedList[i]

		select {
		case client.send <- resp: // Todo: check implications of sending &res to channel and benchmark
		default:
			client.hub.unregister <- client
		}
	}
}

type DeleteMessageEvent struct {
	MessageId  string   `json:"msgId"`      // MessageId or CommentId
	CommentIds []string `json:"commentIds"` // Only used when deleting a top-level message i.e. when MessageId represents a message and not a comment.
}

func (p *DeleteMessageEvent) Handle(e *Event, h *Hub) {

	// Validate
	if h.redis.IsBoardLocked(e.Group) {
		slog.Warn("Cannot delete data in read-only board", "board", e.Group)
		return
	}

	msg, exists := h.redis.GetMessage(p.MessageId)
	if !exists {
		slog.Warn("Message or Comment doesn't exist in DeleteMessageEvent handle", "msgId", p.MessageId)
		return
	}

	if msg.Id != p.MessageId || msg.Group != e.Group {
		slog.Warn("Mismatched message/group in delete event", "msgId", p.MessageId, "group", e.Group)
		return
	}

	isBoardOwner := h.redis.IsBoardOwner(e.Group, e.By)
	canExecute := (msg.By == e.By || isBoardOwner)
	if !canExecute {
		slog.Warn("User not authorized to delete message/comment", "msgId", msg.Id, "user", e.By)
		return
	}

	// Execute
	deleted := false
	if isMessage(msg) {
		commentIds := getValidComments(h, msg, p.CommentIds)
		deleted = h.redis.DeleteMessage(msg.Group, msg.Id, commentIds)
	} else {
		deleted = h.redis.DeleteComment(msg.Group, msg.Id)
	}

	// Publish: to Redis (for broadcasting)
	if deleted {
		h.redis.Publish(msg.Group, &BroadcastArgs{Message: msg, Event: e}) // Todo: Similar to "Like", BroadcastArgs.Message may not be needed here.
	}
}
func (p *DeleteMessageEvent) Broadcast(e *Event, m *Message, h *Hub) {
	// Transform to Outgoing format
	response := m.NewDeleteResponse()

	clients := h.clients[m.Group]
	for client := range clients {
		select {
		case client.send <- response:
		default:
			client.hub.unregister <- client
		}
	}
}

type DeleteAllEvent struct{}

func (p *DeleteAllEvent) Handle(e *Event, h *Hub) {
	// Update Redis
	b, ok := h.redis.GetBoard(e.Group)
	if !ok {
		slog.Warn("Cannot find board when handling DeleteAllEvent", "board", e.Group)
		return
	}
	// validate
	if b.Owner != e.By {
		slog.Warn("Non-owner cannot execute DeleteAllEvent", "board", e.Group, "user", e.By)
		return
	}
	// Delete
	if deleted := h.redis.DeleteAll(b.Id); !deleted {
		slog.Warn("Could not delete all related data for board.", "board", e.Group)
		return
	}
	// Publish to Redis (for broadcasting)
	// *Message is nil as this is not a message related update.
	h.redis.Publish(b.Id, &BroadcastArgs{Message: nil, Event: e})
}
func (p *DeleteAllEvent) Broadcast(e *Event, m *Message, h *Hub) {
	// Transform to Outgoing format
	response := &DeleteAllResponse{Type: "delall"}

	clients := h.clients[e.Group]
	for client := range clients {
		select {
		case client.send <- response:
		default:
			client.hub.unregister <- client
		}
	}
}

type CategoryChangeEvent struct {
	MessageId   string   `json:"msgId"`
	NewCategory string   `json:"newcat"`
	OldCategory string   `json:"oldcat"`
	CommentIds  []string `json:"commentIds"`
}

func (p *CategoryChangeEvent) Handle(e *Event, h *Hub) {
	// Validate
	b, ok := h.redis.GetBoard(e.Group)
	if !ok {
		slog.Warn("Cannot find board when handling CategoryChangeEvent", "board", e.Group)
		return
	}

	if b.Lock {
		slog.Warn("Cannot change message category in read-only board", "board", e.Group)
		return
	}

	msg, exists := h.redis.GetMessage(p.MessageId)
	if !exists {
		slog.Warn("Message doesn't exist in CategoryChangeEvent handle", "msgId", p.MessageId)
		return
	}

	if msg.Id != p.MessageId || msg.Group != e.Group {
		slog.Warn("Mismatched message/group in CategoryChangeEvent handle", "msgId", p.MessageId, "group", e.Group)
		return
	}

	if msg.Category == p.NewCategory {
		slog.Warn("Old and New categories are same. Not changing.", "msgId", p.MessageId, "newCategory", p.NewCategory)
		return
	}

	if !h.redis.IsBoardColumnActive(e.Group, p.NewCategory) {
		slog.Warn("Cannot move message to invalid or inactive category", "board", e.Group, "msgId", p.MessageId, "newCategory", p.NewCategory)
		return
	}

	// Validate before changing category; especially if the message being moved is of the user who created/owns it.
	// Board owner can change category of any message.
	isBoardOwner := b.Owner == e.By
	canExecute := (msg.By == e.By || isBoardOwner)
	if !canExecute {
		slog.Warn("User not authorized to change category message/comment", "msgId", p.MessageId, "user", e.By)
		return
	}

	// Execute
	commentIds := getValidComments(h, msg, p.CommentIds)
	updated := h.redis.UpdateCategory(p.NewCategory, p.MessageId, commentIds)

	// Publish to Redis (for broadcasting)
	// *Message is nil as all message details need not be broadcasted. Event details should be enough.
	if updated {
		h.redis.Publish(msg.Group, &BroadcastArgs{Message: nil, Event: e})
	}
}
func (p *CategoryChangeEvent) Broadcast(e *Event, m *Message, h *Hub) {
	// Transform to Outgoing format
	// We can trust the "p" *CategoryChangeResponse payload here. The Handle must have validated it. Don't want to add another field in BroadcastArgs{}.
	response := &CategoryChangeResponse{Type: "catchng", MessageId: p.MessageId, NewCategory: p.NewCategory}

	clients := h.clients[e.Group]
	for client := range clients {
		select {
		case client.send <- response:
		default:
			client.hub.unregister <- client
		}
	}
}

type TimerEvent struct {
	ExpiryDurationInSeconds uint16 `json:"expiryDurationInSeconds"`
	Stop                    bool   `json:"stop"`
}

func (p *TimerEvent) Handle(e *Event, h *Hub) {
	// The TimerEventHandle handles 2 things separately
	// 	"Stop" - Is used to Stop the timer. When this is true, "ExpiryDurationInSeconds" passed in payload is ignored.
	// 	"ExpiryDurationInSeconds" - Is used to convey the "Start" of a timer.
	// Both are mutually exclusive mostly

	// Validate
	b, ok := h.redis.GetBoard(e.Group)
	if !ok {
		slog.Warn("Cannot find board when handling TimerEvent", "board", e.Group)
		return
	}
	// Validate for both
	if b.Owner != e.By {
		slog.Warn("Non-owner trying to handle TimerEvent", "board", e.Group, "user", e.By)
		return
	}

	nowUnix := time.Now().UTC().Unix()

	// STOP TIMER
	// Validate and Execute for "Stop"
	// Avoid executing "Stop" unnecessarily.
	// Since a "Stop" updates the board's "timerExpiresAtUtc" downsteam, its better to avoid running it when "timerExpiresAtUtc" is 0.
	// "timerExpiresAtUtc" has its initial value of 0. A higher value means we can deduce that a timer has been set atleast once...
	// ...This validation prevents us from loosing that capability to deduce.
	if p.Stop {
		if !timerIsRunning(b.TimerExpiresAtUtc, nowUnix) {
			slog.Warn("Cannot stop timer that isn't running", "board", e.Group)
			return
		}

		if updated := h.redis.StopTimer(b); !updated {
			slog.Warn("Skipping. Unable to update timer during 'Stop' operation.")
			return
		}

		// Publish to Redis (for broadcasting)
		// *Message is nil as this is not a message related update. Timer is a UI gimmick. Find a better way.
		h.redis.Publish(b.Id, &BroadcastArgs{Message: nil, Event: e})
		return
	}

	// START / UPDATE TIMER
	// Validate and Execute for updating timer a.k.a "START"
	if timerIsRunning(b.TimerExpiresAtUtc, nowUnix) {
		slog.Warn("Cannot start Timer again. It is in running state", "board", e.Group)
		return
	}

	// Duration check validation is only when the board owner is trying to set the timer. "Stop" is ignored here.
	if p.ExpiryDurationInSeconds < 1 || p.ExpiryDurationInSeconds > 3600 {
		slog.Warn("Invalid timer duration. Valid duration range is between 1 to 3600 seconds.")
		return
	}

	// Execute Update timer
	if updated := h.redis.UpdateTimer(b, p.ExpiryDurationInSeconds); !updated {
		slog.Warn("Skipping. Unable to update timer information.")
		return
	}

	// Publish to Redis (for broadcasting)
	// *Message is nil as this is not a message related update. Timer is a UI gimmick. Find a better way.
	h.redis.Publish(b.Id, &BroadcastArgs{Message: nil, Event: e})
}
func timerIsRunning(expiresAt, now int64) bool {
	return expiresAt > 0 && expiresAt > now
}
func (p *TimerEvent) Broadcast(e *Event, m *Message, h *Hub) {
	// Transform to Outgoing format
	// redis.Getboard() is called twice in Handle and Broadcast. Let it be that way for now. Don't want to add another field in BroadcastArgs{}.
	board, boardOk := h.redis.GetBoard(e.Group)
	if !boardOk {
		slog.Warn("Cannot find board when broadcasting TimerEvent", "board", e.Group)
		return
	}

	// Prepare timer details
	// remainingTimeInSeconds := board.TimerExpiresAtUtc - time.Now().UTC().Unix()
	// if remainingTimeInSeconds <= 0 {
	// 	remainingTimeInSeconds = 0
	// }
	nowUnix := time.Now().UTC().Unix()
	remainingTimeInSeconds := int64(0)
	if board.TimerExpiresAtUtc > nowUnix {
		remainingTimeInSeconds = board.TimerExpiresAtUtc - nowUnix
	}

	// uint16: This shouldn't error out since we will restrict expiry to max 1 hour (3600 seconds) future time, when saving "board.TimerExpiresAtUtc".
	response := &TimerResponse{Type: "timer", ExpiresInSeconds: uint16(remainingTimeInSeconds)}

	clients := h.clients[e.Group]
	for client := range clients {
		select {
		case client.send <- response:
		default:
			client.hub.unregister <- client
		}
	}
}

type ColumnsChangeEvent struct {
	Columns []*BoardColumn `json:"columns"`
	// Only columns to add/update are sent. Columns to disable aren't sent explicitly.
	// Using same BoardColumn struct that is used for request and redis store. Todo - refactor later.
}

func (p *ColumnsChangeEvent) Handle(e *Event, h *Hub) {
	// validate
	if len(p.Columns) == 0 || len(p.Columns) > 5 {
		slog.Warn("Invalid columns data passed in ColumnsChangeEvent", "board", e.Group)
		return
	}
	for _, col := range p.Columns {
		textLen := utf8.RuneCountInString(col.Text)
		if len(col.Id) > MaxColumnIdSizeBytes || len(col.Color) > MaxColorSizeBytes || textLen > config.Data.MaxCategoryTextLength {
			slog.Warn("Columns info exceeds limit in ColumnsChangeEvent", "col", col.Id, "len", textLen, "len-color", len(col.Color))
			return
		}
	}
	b, ok := h.redis.GetBoard(e.Group)
	if !ok {
		slog.Warn("Cannot find board when handling ColumnsChangeEvent", "board", e.Group)
		return
	}
	if b.Lock {
		slog.Warn("Cannot change columns in read-only board", "board", e.Group)
		return
	}
	if b.Owner != e.By {
		slog.Warn("Non-owner cannot execute ColumnsChangeEvent", "board", e.Group, "user", e.By)
		return
	}
	// Prevent deleting a column with associated messages
	cols, ok := h.redis.GetBoardColumns(b.Id)
	if !ok {
		slog.Warn("Cannot get columns when handling ColumnsChangeEvent", "board", e.Group)
		return
	}
	hasMessages, err := h.redis.HasMessagesForColumnsMarkedForRemoval(b.Id, cols, p.Columns)
	if err != nil {
		slog.Error(err.Error(), "board", e.Group)
		return
	}
	if hasMessages {
		slog.Warn("Cannot reset columns with attached messages in ColumnsChangeEvent", "board", e.Group)
		return
	}

	// execute
	if done := h.redis.ResetBoardColumns(b, cols, p.Columns); !done {
		return
	}

	// Publish to Redis (for broadcasting)
	// *Message is nil as this is not a message related update.
	h.redis.Publish(b.Id, &BroadcastArgs{Message: nil, Event: e})
}
func (p *ColumnsChangeEvent) Broadcast(e *Event, m *Message, h *Hub) {
	// Transform to Outgoing format
	// We can trust the "i" *ColumnsChangeEvent payload here. The Handle must have validated it. Don't want to add another field in BroadcastArgs{}.
	response := &ColumnsChangeResponse{Type: "colreset", BoardColumns: p.Columns}

	clients := h.clients[e.Group]
	for client := range clients {
		select {
		case client.send <- response:
		default:
			client.hub.unregister <- client
		}
	}
}

type TypedEvent struct{}

func (p *TypedEvent) Handle(e *Event, h *Hub) {
	if !config.TypingActivityConfig.Enabled {
		return
	}

	h.redis.Publish(e.Group, &BroadcastArgs{Message: nil, Event: e})
}
func (p *TypedEvent) Broadcast(e *Event, m *Message, h *Hub) {
	response := &TypedResponse{Type: "t", Xid: e.Xid}

	clients := h.clients[e.Group]
	for client := range clients {
		// No need to send response to initiator
		if client.id == e.By {
			continue
		}

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

// Helper: determine if this is a top-level message
func isMessage(msg *Message) bool {
	return msg.ParentId == ""
}

// Helper: return only comments that belong to this message
func getValidComments(h *Hub, msg *Message, ids []string) []string {
	if len(ids) == 0 {
		return nil
	}

	cmts, ok := h.redis.GetMessagesByIds(ids, msg.Group)
	if !ok {
		return nil
	}

	valid := make([]string, 0, len(cmts))
	for _, c := range cmts {
		if c.ParentId == msg.Id {
			valid = append(valid, c.Id)
		}
	}
	return valid
}
