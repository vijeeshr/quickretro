package main

import (
	"encoding/json"
	"log/slog"
	"time"
)

type RegisterEvent struct {
	// The broadcaster of RegisterEvent i.e. "Broadcast" method below sends two types of responses -
	// RegisterResponse: Sent to the client who triggered the RegisterEvent, and
	// UserJoiningResponse: Sent to all the other active clients in the board
	By         string `json:"by"`
	ByNickname string `json:"nickname"`
	Xid        string `json:"xid"`
	Group      string `json:"grp"`
}

func (p *RegisterEvent) Handle(i *Event, h *Hub) {
	// Validate
	if p.By == "" || p.Xid == "" || p.Group == "" {
		return
	}
	// Mostly the board should exist. That check is done during handshake.
	// This is to prevent adding UserPresence data to redis on receipt of a tampered reg event payload with a non-existent board.
	if !h.redis.BoardExists(p.Group) {
		return
	}

	// Execute
	if ok := h.redis.CommitUserPresence(p.Group, &User{Id: p.By, Xid: p.Xid, Nickname: p.ByNickname}); !ok {
		return
	}

	// Publish to Redis (for broadcasting)
	// *Message is nil as this is not a message related update. Find a better way. Generics?
	h.redis.Publish(p.Group, &BroadcastArgs{Message: nil, Event: i})
}
func (i *RegisterEvent) Broadcast(m *Message, h *Hub) {
	// // Todo: Can below calls be pipelined?
	// // Transform to Outgoing format
	// // Don't want to add another field in BroadcastArgs{}.
	// board, boardOk := h.redis.GetBoard(i.Group)
	// if !boardOk {
	// 	slog.Error("No board found in RegisterEvent broadcast", "board", i.Group)
	// 	return
	// }
	// cols, colsOk := h.redis.GetBoardColumns(i.Group)
	// if !colsOk {
	// 	slog.Error("No board columns found in RegisterEvent broadcast", "board", i.Group)
	// 	return
	// }
	// users, usersOk := h.redis.GetUsersPresence(i.Group)
	// if !usersOk {
	// 	slog.Error("Error when getting user presence in RegisterEvent broadcast", "board", i.Group)
	// 	return
	// }
	// messages, messagesOk := h.redis.GetMessages(i.Group)
	// if !messagesOk {
	// 	slog.Error("Error getting messages in RegisterEvent broadcast", "board", i.Group)
	// 	return
	// }
	// comments, commentsOk := h.redis.GetComments(i.Group)
	// if !commentsOk {
	// 	slog.Error("Error getting comments in RegisterEvent broadcast", "board", i.Group)
	// 	return
	// }

	data, ok := h.redis.GetBoardAggregatedData(i.Group)
	if !ok {
		slog.Error("Failed to get board aggregated data", "board", i.Group)
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
	likesInfo, likesOk := h.redis.GetLikesInfo(i.By, ids...)
	if !likesOk {
		slog.Warn("Failed to fetch likes info")
	}
	for in, m := range messages {
		msgRes := m.NewMessageResponse()
		msgRes.Mine = m.By == i.By
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
		cmtRes.Mine = c.By == i.By
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
		// IsBoardOwner:              (i.By == board.Owner),               // TODO: Check if this can be compromised
	}
	// Prepare UserJoiningResponse
	// UserJoiningResponse is sent to all other active clients (except initiator)
	joinResp := UserJoiningResponse{
		Type:     "joining",
		Nickname: i.ByNickname,
		Xid:      i.Xid,
	}

	clients := h.clients[i.Group]

	for client := range clients {
		// // Shallow copy.
		// // Copies the struct value. The fields/slices inside ([]MessageResponse, []UserDetails, []*BoardColumn) are NOT cloned deeply, they still point to the same underlying arrays.
		// // Ensure those fields/slices aren't mutated after being sent from here.
		// response := regResponse
		if client.id == i.By {
			regResponse.IsBoardOwner = client.id == board.Owner // TODO: Check if this is safer than the above, i.By == board.Owner
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
func (p *UserClosingEvent) Handle(i *Event, h *Hub) {
	// Validate
	if p.By == "" || p.Group == "" {
		return
	}

	// Execute
	xidOfRemovedUser, removed := h.redis.RemoveUserPresence(p.Group, p.By)
	// No need to broadcast if there is no xid
	if !removed || xidOfRemovedUser == "" {
		return
	}

	// Populate xid in the event before broadcasting
	p.Xid = xidOfRemovedUser

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
func (i *UserClosingEvent) Broadcast(m *Message, h *Hub) {
	response := &UserClosingResponse{Type: "closing", Xid: i.Xid}

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
	// Validate
	b, ok := h.redis.GetBoard(p.Group)
	if !ok {
		slog.Warn("Cannot find board when handling MaskEvent", "board", p.Group)
		return
	}
	if b.Owner != p.By {
		slog.Warn("Non-owner trying to update board when handling MaskEvent", "board", p.Group, "user", p.By)
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
	h.redis.Publish(b.Id, &BroadcastArgs{Message: nil, Event: i})
}
func (i *MaskEvent) Broadcast(m *Message, h *Hub) {
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
	// validate
	b, ok := h.redis.GetBoard(p.Group)
	if !ok {
		slog.Warn("Cannot find board when handling LockEvent", "board", p.Group)
		return
	}
	if b.Owner != p.By {
		slog.Warn("Non-owner trying to update board when handling LockEvent", "board", p.Group, "user", p.By)
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
	h.redis.Publish(b.Id, &BroadcastArgs{Message: nil, Event: i})
}
func (i *LockEvent) Broadcast(m *Message, h *Hub) {
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
	ByXid      string `json:"byxid"`
	ByNickname string `json:"nickname"`
	Group      string `json:"grp"`
	Content    string `json:"msg"`
	Category   string `json:"cat"`
	ParentId   string `json:"pid"`
	Anonymous  bool   `json:"anon"`
}

func (p *MessageEvent) Handle(i *Event, h *Hub) {
	// Validate
	if h.redis.IsBoardLocked(p.Group) {
		slog.Warn("Cannot save message in read-only board", "board", p.Group)
		return
	}

	msg := p.ToMessage() // From payload

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
		h.redis.Publish(msg.Group, &BroadcastArgs{Message: msg, Event: i})
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

func (i *MessageEvent) Broadcast(m *Message, h *Hub) {
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
	By        string `json:"by"`
	Like      bool   `json:"like"`
}

func (p *LikeMessageEvent) Handle(i *Event, h *Hub) {
	// Validate
	msg, exists := h.redis.GetMessage(p.MessageId) // Todo: Check if fetching a message is needed for a like. Can avoid extra calls. Also BroadcastArgs.Message may not be needed here if removed.
	if !exists {
		slog.Warn("Message doesn't exist in LikeMessageEvent handle", "msgId", p.MessageId)
		return
	}
	// Execute
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
	MessageId  string   `json:"msgId"` // MessageId or CommentId
	By         string   `json:"by"`
	Group      string   `json:"grp"`
	CommentIds []string `json:"commentIds"` // Only used when deleting a top-level message i.e. when MessageId represents a message and not a comment.
}

func (p *DeleteMessageEvent) Handle(i *Event, h *Hub) {

	// Validate
	if h.redis.IsBoardLocked(p.Group) {
		slog.Warn("Cannot delete data in read-only board", "board", p.Group)
		return
	}

	msg, exists := h.redis.GetMessage(p.MessageId)
	if !exists {
		slog.Warn("Message or Comment doesn't exist in DeleteMessageEvent handle", "msgId", p.MessageId)
		return
	}

	if msg.Id != p.MessageId || msg.Group != p.Group {
		slog.Warn("Mismatched message/group in delete event", "msgId", p.MessageId, "group", p.Group)
		return
	}

	isBoardOwner := h.redis.IsBoardOwner(p.Group, p.By)
	canExecute := (msg.By == p.By || isBoardOwner)
	if !canExecute {
		slog.Warn("User not authorized to delete message/comment", "msgId", msg.Id, "user", p.By)
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
		h.redis.Publish(msg.Group, &BroadcastArgs{Message: msg, Event: i}) // Todo: Similar to "Like", BroadcastArgs.Message may not be needed here.
	}
}
func (i *DeleteMessageEvent) Broadcast(m *Message, h *Hub) {
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

type DeleteAllEvent struct {
	By    string `json:"by"`
	Group string `json:"grp"`
}

func (p *DeleteAllEvent) Handle(i *Event, h *Hub) {
	// Update Redis
	b, ok := h.redis.GetBoard(p.Group)
	if !ok {
		slog.Warn("Cannot find board when handling DeleteAllEvent", "board", p.Group)
		return
	}
	// validate
	if b.Owner != p.By {
		slog.Warn("Non-owner cannot execute DeleteAllEvent", "board", p.Group, "user", p.By)
		return
	}
	// Delete
	if deleted := h.redis.DeleteAll(b.Id); !deleted {
		slog.Warn("Could not delete all related data for board.", "board", p.Group)
		return
	}
	// Publish to Redis (for broadcasting)
	// *Message is nil as this is not a message related update.
	h.redis.Publish(b.Id, &BroadcastArgs{Message: nil, Event: i})
}
func (i *DeleteAllEvent) Broadcast(m *Message, h *Hub) {
	// Transform to Outgoing format
	response := &DeleteAllResponse{Type: "delall"}

	clients := h.clients[i.Group]
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
	By          string   `json:"by"`
	Group       string   `json:"grp"`
	NewCategory string   `json:"newcat"`
	OldCategory string   `json:"oldcat"`
	CommentIds  []string `json:"commentIds"`
}

func (p *CategoryChangeEvent) Handle(i *Event, h *Hub) {
	// Validate
	if h.redis.IsBoardLocked(p.Group) {
		slog.Warn("Cannot change message category in read-only board", "board", p.Group)
		return
	}

	msg, exists := h.redis.GetMessage(p.MessageId)
	if !exists {
		slog.Warn("Message doesn't exist in CategoryChangeEvent handle", "msgId", p.MessageId)
		return
	}

	if msg.Id != p.MessageId || msg.Group != p.Group {
		slog.Warn("Mismatched message/group in CategoryChangeEvent handle", "msgId", p.MessageId, "group", p.Group)
		return
	}

	if msg.Category == p.NewCategory {
		slog.Warn("Old and New categories are same. Not changing.", "msgId", p.MessageId, "newCategory", p.NewCategory)
		return
	}

	// Validate before changing category; especially if the message being moved is of the user who created/owns it.
	// Board owner can change category of any message.
	isBoardOwner := h.redis.IsBoardOwner(p.Group, p.By)
	canExecute := (msg.By == p.By || isBoardOwner)
	if !canExecute {
		slog.Warn("User not authorized to change category message/comment", "msgId", p.MessageId, "user", p.By)
		return
	}

	// Execute
	commentIds := getValidComments(h, msg, p.CommentIds)
	updated := h.redis.UpdateCategory(p.NewCategory, p.MessageId, commentIds)

	// Publish to Redis (for broadcasting)
	// *Message is nil as all message details need not be broadcasted. Event details should be enough.
	if updated {
		h.redis.Publish(msg.Group, &BroadcastArgs{Message: nil, Event: i})
	}
}
func (i *CategoryChangeEvent) Broadcast(m *Message, h *Hub) {
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

	// Validate
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
func (i *TimerEvent) Broadcast(m *Message, h *Hub) {
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

type ColumnsChangeEvent struct {
	By      string         `json:"by"`
	Group   string         `json:"grp"`
	Columns []*BoardColumn `json:"columns"`
	// Only columns to add/update are sent. Columns to disable aren't sent explicitly.
	// Using same BoardColumn struct that is used for request and redis store. Todo - refactor later.
}

func (p *ColumnsChangeEvent) Handle(i *Event, h *Hub) {
	// validate
	if len(p.Columns) == 0 || len(p.Columns) > 5 {
		slog.Warn("Invalid columns data passed in ColumnsChangeEvent", "board", p.Group)
		return
	}
	for _, col := range p.Columns {
		if len([]rune(col.Text)) > config.Server.MaxCategoryTextLength {
			slog.Warn("Columns text length exceeds limit in create board request payload", "len", len([]rune(col.Text)), "col", col.Id)
			return
		}
	}
	b, ok := h.redis.GetBoard(p.Group)
	if !ok {
		slog.Warn("Cannot find board when handling ColumnsChangeEvent", "board", p.Group)
		return
	}
	if b.Lock {
		slog.Warn("Cannot change columns in read-only board", "board", p.Group)
		return
	}
	if b.Owner != p.By {
		slog.Warn("Non-owner cannot execute ColumnsChangeEvent", "board", p.Group, "user", p.By)
		return
	}
	// Prevent deleting a column with associated messages
	cols, ok := h.redis.GetBoardColumns(b.Id)
	if !ok {
		slog.Warn("Cannot get columns when handling ColumnsChangeEvent", "board", p.Group)
		return
	}
	hasMessages, err := h.redis.HasMessagesForColumnsMarkedForRemoval(b.Id, cols, p.Columns)
	if err != nil {
		slog.Error(err.Error(), "board", p.Group)
		return
	}
	if hasMessages {
		slog.Warn("Cannot reset columns with attached messages in ColumnsChangeEvent", "board", p.Group)
		return
	}

	// execute
	if done := h.redis.ResetBoardColumns(b, cols, p.Columns); !done {
		return
	}

	// Publish to Redis (for broadcasting)
	// *Message is nil as this is not a message related update.
	h.redis.Publish(b.Id, &BroadcastArgs{Message: nil, Event: i})
}
func (i *ColumnsChangeEvent) Broadcast(m *Message, h *Hub) {
	// Transform to Outgoing format
	// We can trust the "i" *ColumnsChangeEvent payload here. The Handle must have validated it. Don't want to add another field in BroadcastArgs{}.
	response := &ColumnsChangeResponse{Type: "colreset", BoardColumns: i.Columns}

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
