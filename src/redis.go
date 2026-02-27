package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConnector struct {
	ctx        context.Context
	client     *redis.Client
	subscriber *redis.PubSub
	timeToLive time.Duration
}

func NewRedisConnector(ctx context.Context, timeToLive time.Duration) *RedisConnector {

	// rdb := redis.NewClient(&redis.Options{
	// 	Addr:     "localhost:6379",
	// 	Password: "", // no password set
	// 	DB:       0,  // use default DB
	// })

	opt, err := redis.ParseURL(envConfig.RedisConnStr)
	if err != nil {
		slog.Error("Cannot parse Redis connection string", "err", err)
	}

	// Todo: Add auth and pull from config
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		// Addrs:    []string{":6379"},
		// Addrs:    []string{"my-redis:6379"},
		Addrs:    []string{opt.Addr},
		Username: opt.Username,
		Password: opt.Password,
		DB:       opt.DB,
	})

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		slog.Error("Cannot connect to Redis", "err", err)
		os.Exit(1)
	}

	// return rdb.(*redis.Client), rdb.(*redis.Client).Subscribe(ctx)
	return &RedisConnector{client: rdb.(*redis.Client), subscriber: rdb.(*redis.Client).Subscribe(ctx), timeToLive: timeToLive, ctx: ctx}
}

func (c *RedisConnector) Subscribe(redisChannel ...string) {
	if err := c.subscriber.Subscribe(c.ctx, redisChannel...); err != nil {
		slog.Error("Unable to subscribe", "err", err, "channels", redisChannel)
	}
}

func (c *RedisConnector) Unsubscribe(redisChannel ...string) {
	if err := c.subscriber.Unsubscribe(c.ctx, redisChannel...); err != nil {
		slog.Error("Unable to Unsubscribe", "err", err, "channels", redisChannel)
	}
}

func (c *RedisConnector) Publish(redisChannel string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Marshal error on publish", "err", err, "channel", redisChannel)
	}
	if err := c.client.Publish(c.ctx, redisChannel, data).Err(); err != nil {
		slog.Error("Publish error", "err", err, "channel", redisChannel)
	}
}

func (c *RedisConnector) CreateBoard(b *Board, cols []*BoardColumn) bool {
	key := boardKey(b.Id)
	boardColsKey := boardColsKey(b.Id) // Boardwise-ColIds

	currentTime := time.Now().UTC()
	currentTimeUtcSeconds := currentTime.Unix()
	autoDeleteTime := currentTime.Add(c.timeToLive)
	autoDeleteTimeUtcSeconds := autoDeleteTime.Unix()

	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(c.ctx, key,
			"id", b.Id,
			"name", b.Name,
			"team", b.Team,
			"owner", b.Owner,
			"status", int(b.Status),
			"mask", b.Mask,
			"lock", b.Lock,
			"createdAtUtc", currentTimeUtcSeconds,
			"autoDeleteAtUtc", autoDeleteTimeUtcSeconds,
		)
		// Columns
		for _, col := range cols {
			colKey := boardColKey(b.Id, col.Id)
			pipe.HSet(c.ctx, colKey,
				"id", col.Id,
				"text", col.Text,
				"isDefault", col.IsDefault,
				"color", col.Color,
				"pos", col.Position,
			)
			pipe.ExpireAt(c.ctx, colKey, autoDeleteTime)
			pipe.SAdd(c.ctx, boardColsKey, col.Id)
		}
		pipe.ExpireAt(c.ctx, boardColsKey, autoDeleteTime)
		pipe.ExpireAt(c.ctx, key, autoDeleteTime)
		return nil
	})

	if err != nil {
		slog.Error("Failed to create board in Redis", "err", err, "board", b, "cols", cols)
		return false
	}

	return true
}

func (c *RedisConnector) HasMessagesForColumnsMarkedForRemoval(boardId string, oldCols []*BoardColumn, newCols []*BoardColumn) (bool, error) {
	// Compare oldCols vs newCols to find colIds that existed before but no longer exist.
	oldMap := make(map[string]struct{}, len(oldCols))
	newMap := make(map[string]struct{}, len(newCols))
	for _, c := range oldCols {
		oldMap[c.Id] = struct{}{}
	}
	for _, c := range newCols {
		newMap[c.Id] = struct{}{}
	}

	var missingIds []string // missingIds are ids of columns marked for removal
	for oldId := range oldMap {
		if _, stillExists := newMap[oldId]; !stillExists {
			missingIds = append(missingIds, oldId)
		}
	}

	if len(missingIds) == 0 {
		return false, nil
	}

	// Fetch messages
	messages, ok := c.GetMessages(boardId)
	if !ok {
		return false, fmt.Errorf("unable to fetch messages for board %s", boardId)
	}

	// Convert list to a set for quick lookup
	disableSet := make(map[string]struct{}, len(missingIds))
	for _, id := range missingIds {
		disableSet[id] = struct{}{}
	}

	// Check messages: any message belonging to a disabled column?
	for _, m := range messages {
		if _, exists := disableSet[m.Category]; exists {
			return true, nil
		}
	}

	return false, nil
}

func (c *RedisConnector) ResetBoardColumns(b *Board, oldCols []*BoardColumn, newCols []*BoardColumn) bool {
	boardColsKey := boardColsKey(b.Id)
	autoDeleteTime := time.Unix(b.AutoDeleteAtUtc, 0).UTC()

	// Convert to map for quick lookup
	oldMap := make(map[string]*BoardColumn, len(oldCols))
	newMap := make(map[string]*BoardColumn, len(newCols))
	for _, c := range oldCols {
		oldMap[c.Id] = c
	}
	for _, c := range newCols {
		newMap[c.Id] = c
	}

	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {

		// DELETE columns that no longer exist
		for oldId := range oldMap {
			if _, stillExists := newMap[oldId]; !stillExists {
				colKey := boardColKey(b.Id, oldId)
				pipe.Del(c.ctx, colKey)
			}
		}

		// UPSERT (create or deep-update) columns
		for _, newCol := range newCols {

			oldCol, existed := oldMap[newCol.Id]
			colKey := boardColKey(b.Id, newCol.Id)

			// If didn’t exist → full create
			if !existed {
				pipe.HSet(c.ctx, colKey,
					"id", newCol.Id,
					"text", newCol.Text,
					"isDefault", newCol.IsDefault,
					"color", newCol.Color,
					"pos", newCol.Position,
				)
				pipe.ExpireAt(c.ctx, colKey, autoDeleteTime)
				continue
			}

			// Deep diff: Only update changed fields
			var changes []any

			if oldCol.Text != newCol.Text {
				changes = append(changes, "text", newCol.Text)
			}
			if oldCol.Color != newCol.Color {
				changes = append(changes, "color", newCol.Color)
			}
			if oldCol.IsDefault != newCol.IsDefault {
				changes = append(changes, "isDefault", newCol.IsDefault)
			}
			if oldCol.Position != newCol.Position {
				changes = append(changes, "pos", newCol.Position)
			}

			if len(changes) > 0 {
				pipe.HSet(c.ctx, colKey, changes...)
			}

			// Note: TTL is preserved automatically unless created new
		}

		// Rebuild the column ID set
		pipe.Del(c.ctx, boardColsKey)
		for _, col := range newCols {
			pipe.SAdd(c.ctx, boardColsKey, col.Id)
		}
		pipe.ExpireAt(c.ctx, boardColsKey, autoDeleteTime)

		return nil
	})

	if err != nil {
		slog.Error("Failed to deep diff reset columns", "err", err, "board", b.Id)
		return false
	}

	return true
}

func (c *RedisConnector) UpdateMasking(b *Board, mask bool) bool {
	// Todo: Deduplicate with UpdateBoardLock() & UpdateTimer()
	key := boardKey(b.Id)
	if _, err := c.client.HSet(c.ctx, key, "mask", mask).Result(); err != nil {
		slog.Error("Failed to mask/unmask", "err", err, "board", b)
		return false
	}
	return true
}

func (c *RedisConnector) UpdateBoardLock(b *Board, lock bool) bool {
	// Todo: Deduplicate with UpdateMasking() & UpdateTimer()
	key := boardKey(b.Id)
	if _, err := c.client.HSet(c.ctx, key, "lock", lock).Result(); err != nil {
		slog.Error("Failed to lock/unlock", "err", err, "board", b)
		return false
	}
	return true
}

func (c *RedisConnector) UpdateTimer(b *Board, expiryDurationInSeconds uint16) bool {
	// Todo: Deduplicate with UpdateMasking() & UpdateBoardLock()
	key := boardKey(b.Id)
	duration := time.Duration(expiryDurationInSeconds) * time.Second
	expiryTime := time.Now().UTC().Add(duration).Unix()

	if _, err := c.client.HSet(c.ctx, key, "timerExpiresAtUtc", expiryTime).Result(); err != nil {
		slog.Error("Failed to update board timer", "err", err, "board", b)
		return false
	}
	return true
}

func (c *RedisConnector) StopTimer(b *Board) bool {
	key := boardKey(b.Id)
	// To stop the timer, just reset the timerExpiresAtUtc to one second before current time.
	// This will cause the expiryTimeInSeconds to be sent as 0 (if expiryTime - curentTime is negative, its also sent as zero)
	// ...e.g check TimerEvent.broadcast, RegEvent.broadcast
	expiryTime := time.Now().UTC().Unix() - 1

	if _, err := c.client.HSet(c.ctx, key, "timerExpiresAtUtc", expiryTime).Result(); err != nil {
		slog.Error("Failed to update board timer during a 'Stop'", "err", err, "board", b)
		return false
	}
	return true
}

func (c *RedisConnector) BoardExists(boardId string) bool {
	key := boardKey(boardId)

	k, err := c.client.Exists(c.ctx, key).Result()
	if err != nil {
		slog.Error("Cannot find board in Redis", "err", err, "boardId", boardId)
		return false
	}
	if k != 1 {
		slog.Warn("Non-existent board", "boardId", boardId)
		return false
	}

	return true
}

func (c *RedisConnector) GetBoard(boardId string) (*Board, bool) {
	var b Board
	key := boardKey(boardId)

	if err := c.client.HGetAll(c.ctx, key).Scan(&b); err != nil {
		slog.Error("Failed to get board from Redis", "err", err, "boardId", boardId)
		return nil, false
	}
	// Assuming Id as empty to decide the key doesn't exist. This is done to avoid an additional EXISTS call to Redis.
	if b.Id == "" {
		return nil, false
	}

	return &b, true
}

func (c *RedisConnector) IsBoardOwner(boardId string, userId string) bool {
	key := boardKey(boardId)

	if userId == "" || boardId == "" {
		return false
	}

	owner, err := c.client.HGet(c.ctx, key, "owner").Result()
	if err != nil {
		slog.Error("Cannot find board in Redis", "err", err, "boardId", boardId)
		return false
	}

	return userId == owner
}

func (c *RedisConnector) IsBoardLocked(boardId string) bool {
	if boardId == "" {
		return true
	}

	key := boardKey(boardId)
	isLocked, err := c.client.HGet(c.ctx, key, "lock").Result()
	if err != nil {
		slog.Error("Cannot find board in Redis", "err", err, "boardId", boardId)
		return true
	}

	return isLocked == "1"
}

func (c *RedisConnector) IsBoardColumnActive(boardId, colId string) bool {
	key := boardColsKey(boardId)

	ok, err := c.client.SIsMember(c.ctx, key, colId).Result()
	if err != nil {
		slog.Error("Failed to check board column membership", "err", err, "board", boardId, "col", colId)
		return false
	}

	return ok
}

func (c *RedisConnector) GetBoardColumns(boardId string) ([]*BoardColumn, bool) {
	cols := make([]*BoardColumn, 0)

	key := boardColsKey(boardId)
	colIds, err := c.client.SMembers(c.ctx, key).Result()
	if err != nil {
		slog.Error("Failed to get columns from Redis", "err", err, "boardId", boardId)
		return cols, false
	}

	cmds, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		for _, id := range colIds {
			key := boardColKey(boardId, id)
			pipe.HGetAll(c.ctx, key)
		}
		return nil
	})
	if err != nil {
		return cols, false
	}

	for _, cmd := range cmds {
		var c BoardColumn
		if err := cmd.(*redis.MapStringStringCmd).Scan(&c); err != nil {
			slog.Error("Failed to get/map column definition from Redis", "err", err, "boardId", boardId)
			continue
		}
		cols = append(cols, &c)
	}

	return cols, true
}

func (c *RedisConnector) EnsureUser(boardId, userId, nickname string) (*User, bool) {
	userKey := boardUserKey(boardId, userId)
	xidSeqKey := boardUserXidKey(boardId)

	// Try to load existing user
	var user User
	if err := c.client.HGetAll(c.ctx, userKey).Scan(&user); err != nil {
		slog.Error("Failed to HGETALL user", "err", err, "boardId", boardId, "userId", userId)
		return nil, false
	}

	// User already exists
	if user.Id != "" && user.Xid != "" {
		// Update nickname if changed
		// Nickname update is best-effort; failure should not block upstream call
		if user.Nickname != nickname {
			if err := c.client.HSet(c.ctx, userKey, "nickname", nickname).Err(); err != nil {
				slog.Error("Failed updating nickname", "err", err, "boardId", boardId, "userId", userId)
			}
			user.Nickname = nickname
		}

		return &user, true
	}

	// Create new user
	var xidCmd *redis.IntCmd

	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		xidCmd = pipe.Incr(c.ctx, xidSeqKey)
		pipe.Expire(c.ctx, xidSeqKey, c.timeToLive) // Todo: We can try to expire this earlier by looking at Board.AutoDeleteAtUtc. But the requires a call to get board details. Skipping it for now.

		pipe.HSet(c.ctx, userKey,
			"id", userId,
			"nickname", nickname,
		)
		pipe.Expire(c.ctx, userKey, c.timeToLive)
		return nil
	})

	if err != nil {
		slog.Error("Failed creating user", "err", err, "boardId", boardId, "userId", userId)
		return nil, false
	}

	// Persist xid (now that we know it)
	xid := strconv.FormatInt(xidCmd.Val(), 10)

	if err := c.client.HSet(c.ctx, userKey, "xid", xid).Err(); err != nil {
		// Todo: Should the above key be deleted on failure to create xid?
		slog.Error("Failed setting xid", "err", err, "boardId", boardId, "userId", userId)
		return nil, false
	}

	user = User{
		Id:       userId,
		Xid:      xid,
		Nickname: nickname,
	}

	return &user, true
}

func (c *RedisConnector) CommitUserPresence(boardId string, userId string) bool {
	boardUsersKey := boardUsersPresenceKey(boardId)

	// Todo: Remove pipeline?
	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		pipe.SAdd(c.ctx, boardUsersKey, userId)
		pipe.Expire(c.ctx, boardUsersKey, c.timeToLive) // Todo: We can try to expire this earlier by looking at Board.AutoDeleteAtUtc. But the requires a call to get board details. Skipping it for now.
		return nil
	})

	if err != nil {
		slog.Error("Failed committing user presence to Redis", "err", err, "boardId", boardId, "user", userId)
		return false
	}

	return true
}

func (c *RedisConnector) RemoveUserPresence(boardId string, userId string) bool {
	boardUsersKey := boardUsersPresenceKey(boardId)

	// removed, err := c.client.SRem(c.ctx, boardUsersKey, userId).Result()
	// if err != nil {
	// 	return false
	// }
	// if removed == 0 {
	// 	slog.Debug("User presence already absent", "boardId", boardId, "userId", userId)
	// }

	if err := c.client.SRem(c.ctx, boardUsersKey, userId).Err(); err != nil {
		slog.Error("Failed removing user presence from Redis", "err", err, "boardId", boardId, "userId", userId)
		return false
	}

	return true
}

// Deprecated: No longer used
func (c *RedisConnector) GetUsersPresence(boardId string) ([]*User, bool) {
	users := make([]*User, 0)

	key := boardUsersPresenceKey(boardId)
	userIds, err := c.client.SMembers(c.ctx, key).Result()
	if err != nil {
		slog.Error("Failed getting userIds from Redis", "err", err, "boardId", boardId)
		return users, false
	}

	cmds, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		for _, id := range userIds {
			key := boardUserKey(boardId, id)
			pipe.HGetAll(c.ctx, key)
		}
		return nil
	})
	if err != nil {
		slog.Error("Error in GetUsersPresence", "err", err)
		return users, false
	}

	for _, cmd := range cmds {
		var u User
		if err := cmd.(*redis.MapStringStringCmd).Scan(&u); err != nil {
			slog.Error("Failed getting/mapping users from Redis", "err", err, "boardId", boardId)
			continue
		}
		users = append(users, &u)
	}

	return users, true
}

// Deprecated: No longer used
func (c *RedisConnector) GetPresentUserIds(boardId string) ([]string, bool) {
	key := boardUsersPresenceKey(boardId)
	ids, err := c.client.SMembers(c.ctx, key).Result()
	if err != nil {
		slog.Error("Failed getting userIds from Redis", "err", err, "boardId", boardId)
		return ids, false
	}
	return ids, true
}

func (c *RedisConnector) GetMessage(msgId string) (*Message, bool) {
	var message Message
	key := msgKey(msgId)

	if err := c.client.HGetAll(c.ctx, key).Scan(&message); err != nil {
		slog.Error("Failed getting/mapping message from Redis", "err", err, "msgId", msgId)
		return nil, false
	}
	// Assuming Id as empty to decide the key doesn't exist. This is done to avoid an additional EXISTS call to Redis.
	if message.Id == "" {
		return nil, false
	}

	return &message, true
}

func (c *RedisConnector) GetMessages(boardId string) ([]*Message, bool) {
	key := boardMsgsKey(boardId)
	messageIds, err := c.client.SMembers(c.ctx, key).Result()
	if err != nil {
		slog.Error("Failed getting messageIds from Redis", "err", err, "boardId", boardId)
		return make([]*Message, 0), false
	}
	return c.GetMessagesByIds(messageIds, boardId)
}

// Deprecated: No longer used
func (c *RedisConnector) GetComments(boardId string) ([]*Message, bool) {
	key := boardCmtsKey(boardId)
	commentIds, err := c.client.SMembers(c.ctx, key).Result()
	if err != nil {
		slog.Error("Failed getting commentIds from Redis", "err", err, "boardId", boardId)
		return make([]*Message, 0), false
	}
	return c.GetMessagesByIds(commentIds, boardId)
}

func (c *RedisConnector) GetMessagesByIds(ids []string, boardId string) ([]*Message, bool) {
	messages := make([]*Message, 0)

	if len(ids) == 0 {
		return messages, true
	}

	cmds, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		for _, id := range ids {
			key := msgKey(id)
			pipe.HGetAll(c.ctx, key)
		}
		return nil
	})
	if err != nil {
		return messages, false
	}

	for _, cmd := range cmds {
		var m Message
		if err := cmd.(*redis.MapStringStringCmd).Scan(&m); err != nil {
			slog.Error("Failed getting/mapping message from Redis", "err", err, "boardId", boardId)
			continue
		}
		messages = append(messages, &m)
	}

	return messages, true
}

func (c *RedisConnector) GetLikesCount(msgId string) int64 {
	key := msgLikesKey(msgId)

	count, err := c.client.SCard(c.ctx, key).Result()
	if err != nil {
		slog.Error("Failed getting likes count from Redis", "err", err, "msgId", msgId)
		return 0
	}
	return count
}

func (c *RedisConnector) HasLiked(msgId string, users []string) []bool {
	key := msgLikesKey(msgId)

	results, err := c.client.SMIsMember(c.ctx, key, users).Result()
	if err != nil {
		// default to all false
		return make([]bool, len(users))
	}

	return results // []bool matching the order of users
}

// Store helper - DTO
type LikeInfo struct {
	Count int64
	Liked bool
}

func (c *RedisConnector) GetLikesInfo(by string, msgIds ...string) (map[string]LikeInfo, bool) {
	result := make(map[string]LikeInfo, len(msgIds))
	if len(msgIds) == 0 {
		return result, true
	}

	// We'll store both IntCmds (for count) and BoolCmds (for liked)
	type likeCmds struct {
		count *redis.IntCmd
		liked *redis.BoolCmd
	}
	cmds := make(map[string]likeCmds, len(msgIds))

	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		for _, id := range msgIds {
			key := msgLikesKey(id)
			countCmd := pipe.SCard(c.ctx, key)
			likedCmd := pipe.SIsMember(c.ctx, key, by)
			cmds[id] = likeCmds{count: countCmd, liked: likedCmd}
		}
		return nil
	})
	if err != nil {
		return nil, false
	}

	// Trusting Redis pipeline execution order for populating the result
	for id, cmdPair := range cmds {
		count, err1 := cmdPair.count.Result()
		if err1 != nil {
			count = 0
		}

		liked, err2 := cmdPair.liked.Result()
		if err2 != nil {
			liked = false
		}

		result[id] = LikeInfo{
			Count: count,
			Liked: liked,
		}
	}

	return result, true
}

func (c *RedisConnector) Save(msg *Message, modes ...SaveMode) bool {
	key := msgKey(msg.Id)
	messagesKey := boardMsgsKey(msg.Group)
	commentsKey := boardCmtsKey(msg.Group)

	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		// Always save the message/comment to the Hash
		pipe.HSet(c.ctx, key,
			"id", msg.Id,
			"by", msg.By,
			"byxid", msg.ByXid,
			"nickname", msg.ByNickname,
			"group", msg.Group,
			"content", msg.Content,
			"category", msg.Category,
			"anon", msg.Anonymous,
			"pid", msg.ParentId,
		)
		pipe.Expire(c.ctx, key, c.timeToLive) // Todo: We can try to expire this earlier by looking at Board.AutoDeleteAtUtc. But requires a call to get board details. Skipping it for now.

		// Handle optional extra behavior
		if len(modes) > 0 {
			// Only considering first variadic argument in modes ...SaveMode
			switch modes[0] {
			case AsNewMessage:
				// Add Id to board:msg:{boardId} SET
				pipe.SAdd(c.ctx, messagesKey, msg.Id)         // This is safe to be called multiple times too, without adding new entries to the Set.
				pipe.Expire(c.ctx, messagesKey, c.timeToLive) // Todo: We can try to expire this earlier by looking at Board.AutoDeleteAtUtc. But requires a call to get board details. Skipping it for now.

			case AsNewComment:
				// Add Id to board:cmts:{boardId} SET
				pipe.SAdd(c.ctx, commentsKey, msg.Id)         // This is safe to be called multiple times too, without adding new entries to the Set.
				pipe.Expire(c.ctx, commentsKey, c.timeToLive) // Todo: We can try to expire this earlier by looking at Board.AutoDeleteAtUtc. But requires a call to get board details. Skipping it for now.
			}
		}

		return nil
	})

	if err != nil {
		slog.Error("Failed to completely save message/comment to Redis", "err", err, "payloads", msg, "modes", modes)
		return false
	}

	return true
}

func (c *RedisConnector) Like(msgId string, by string, like bool) bool {
	key := msgLikesKey(msgId)

	var affected int64
	if like {
		cmds, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
			pipe.SAdd(c.ctx, key, by)
			pipe.Expire(c.ctx, key, c.timeToLive)
			return nil
		})
		if err != nil {
			slog.Error("Error when liking", "err", err, "msgId", msgId, "by", by)
			return false
		}
		affected, _ = cmds[0].(*redis.IntCmd).Result()
	} else {
		result, err := c.client.SRem(c.ctx, key, by).Result()
		if err != nil {
			slog.Error("Error when unliking", "err", err, "msgId", msgId, "by", by)
			return false
		}
		affected = result
	}
	if affected == 0 {
		if like {
			slog.Warn("Cannot like a message which is already liked", "msgId", msgId, "by", by, "like", like)
		} else {
			slog.Warn("Message must be liked for it to be unliked", "msgId", msgId, "by", by, "like", like)
		}
		return false
	}
	return true
}

func (c *RedisConnector) DeleteMessage(group string, msgId string, commentIds []string) bool {
	/*
		DELETE SINGLE MESSAGE
		---------------------
		1. Delete HASH msg:{messageId}
		2. Delete SET msg:likes:{messageId}
		3. Remove messageId entry from SET board:msg:{boardId}
		4. For each associated comment:
			4.1. Delete HASH msg:{commentId}
			4.2. Remove commentId entry from SET board:cmts:{boardId}
	*/
	key := msgKey(msgId)
	likesKey := msgLikesKey(msgId)
	messagesKey := boardMsgsKey(group)
	commentsKey := boardCmtsKey(group)

	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		// Delete the top-level message and other related data
		pipe.Del(c.ctx, key, likesKey)
		pipe.SRem(c.ctx, messagesKey, msgId)
		for _, cid := range commentIds {
			cKey := msgKey(cid)
			// Delete each attached comment
			// Comments don't have likes right now
			pipe.Del(c.ctx, cKey)
			// Remove comment reference from board-level comments list
			pipe.SRem(c.ctx, commentsKey, cid)
		}
		return nil
	})

	// Todo: Should individual results be checked from the pipeline response?
	// Todo: Look into TxPipeline
	if err != nil {
		slog.Error("Error deleting data from DeleteMessage", "msgId", msgId, "commentIds", commentIds, "err", err)
		return false
	}

	return true
}

func (c *RedisConnector) DeleteComment(group string, commentId string) bool {
	/*
		DELETE SINGLE COMMENT
		---------------------
		1. Delete HASH msg:{messageId}
		2. Remove messageId entry from SET board:cmts:{boardId}
	*/
	key := msgKey(commentId)
	commentsKey := boardCmtsKey(group)

	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		// Delete the comment data
		// Comments don't have likes right now
		pipe.Del(c.ctx, key)
		// Remove comment reference from board-level comments list
		pipe.SRem(c.ctx, commentsKey, commentId)
		return nil
	})

	// Todo: Should individual results be checked from the pipeline response?
	// Todo: Look into TxPipeline
	if err != nil {
		slog.Error("Error deleting comment from DeleteComment", "commentId", commentId, "err", err)
		return false
	}

	return true
}

func (c *RedisConnector) DeleteAll(boardId string) bool {
	/*
		Board
		(KEY)board:{boardId}					(VALUE)board					Board - Redis Hash. The board details. Owned by single user.

		Messages, Comments, Likes
		(KEY)board:msg:{boardId}				(VALUE)[messageIds] 			Board-wise Messages - Redis Set. Useful for fetching list of messages.
		(KEY)board:cmts:{boardId}      			(VALUE)[commentIds]      		Board-wise Comments - Redis Set. For fetching all comments.
		(KEY)msg:{messageId}					(VALUE)message					Message - Redis Hash. Useful for fetch/add/update for an individual message.
		(KEY)msg:likes:{messageId}				(VALUE)[userIds]				Likes - Redis Set. For recording likes/votes for a message

		Users
		(KEY)board:presence:{boardId}			(VALUE)[userIds]				Board-wise Live(Connected) Users - Redis Set.
		(KEY)board:user:{boardId}:{userId}	    (VALUE)User						User - Redis Hash. User master. Keeping as board specific.
		(KEY)board:user:xid:seq:{boardId}		(VALUE)last_xid					Last generated sequential xid for Board - Redis INCR. Used to generate sequential Xids.

		Columns
		(KEY)board:col:{boardId}				(Value)[colIds]					Board-wise columns - Redis Set. Just a list of colIds for a board.
		(KEY)board:col:{boardId}:{colId}		(VALUE)column					Column - Redis Hash. Column definition for a Board.
	*/
	ctx := c.ctx

	boardMsgsKey := boardMsgsKey(boardId)
	boardCommsKey := boardCmtsKey(boardId)
	boardUsersKey := boardUsersPresenceKey(boardId)
	boardUserXidKey := boardUserXidKey(boardId)
	boardColsKey := boardColsKey(boardId)
	boardKey := boardKey(boardId)

	// Collect all message Ids, comment Ids, user Ids, and column Ids
	// Pipeline SMEMBERS (read phase)
	readPipe := c.client.Pipeline()
	msgsCmd := readPipe.SMembers(ctx, boardMsgsKey)
	cmtsCmd := readPipe.SMembers(ctx, boardCommsKey)
	usrsCmd := readPipe.SMembers(ctx, boardUsersKey)
	colsCmd := readPipe.SMembers(ctx, boardColsKey)

	if _, err := readPipe.Exec(ctx); err != nil {
		slog.Error("Redis SMEMBERS pipeline failed in DeleteAll", "boardId", boardId, "err", err)
		return false
	}

	messageIds := msgsCmd.Val()
	commentIds := cmtsCmd.Val()
	userIds := usrsCmd.Val()
	colIds := colsCmd.Val()

	// Pipeline Deletes (write phase)
	_, err := c.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {

		// Delete messages + likes
		for _, msgId := range messageIds {
			likesKey := msgLikesKey(msgId)
			msgKey := msgKey(msgId)
			pipe.Del(ctx, likesKey, msgKey)
		}
		pipe.Del(ctx, boardMsgsKey)

		// Delete comments
		for _, cid := range commentIds {
			commentKey := msgKey(cid)
			pipe.Del(ctx, commentKey)
		}
		pipe.Del(ctx, boardCommsKey)

		// Delete users
		for _, userId := range userIds {
			userKey := boardUserKey(boardId, userId)
			pipe.Del(ctx, userKey)
		}
		pipe.Del(ctx, boardUsersKey)
		pipe.Del(ctx, boardUserXidKey)

		// Delete columns
		for _, colId := range colIds {
			colKey := boardColKey(boardId, colId)
			pipe.Del(ctx, colKey)
		}
		pipe.Del(ctx, boardColsKey)

		// Delete board hash
		pipe.Del(ctx, boardKey)

		return nil
	})

	if err != nil {
		slog.Error("Error when deleting all board data from Redis", "boardId", boardId, "err", err)
		return false
	}

	return true
}

func (c *RedisConnector) UpdateCategory(category string, msgId string, commentIds []string) bool {
	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		// Update main message
		key := msgKey(msgId)
		pipe.HSet(c.ctx, key, "category", category)
		// Update associated comments (if any)
		for _, cid := range commentIds {
			ckey := msgKey(cid)
			pipe.HSet(c.ctx, ckey, "category", category)
		}
		return nil
	})

	if err != nil {
		slog.Error("Failed to update category", "err", err, "category", category, "msgId", msgId, "commentIds", commentIds)
		return false
	}
	return true
}

// Store helper - DTO
type BoardAggregatedData struct {
	Board    *Board
	Columns  []*BoardColumn
	Users    []*User
	Messages []*Message
	Comments []*Message
}

func (c *RedisConnector) GetBoardAggregatedData(boardId string) (*BoardAggregatedData, bool) {
	pipe := c.client.Pipeline()

	boardCmd := pipe.HGetAll(c.ctx, boardKey(boardId))
	colIdsCmd := pipe.SMembers(c.ctx, boardColsKey(boardId))
	userIdsCmd := pipe.SMembers(c.ctx, boardUsersPresenceKey(boardId))
	msgIdsCmd := pipe.SMembers(c.ctx, boardMsgsKey(boardId))
	cmtIdsCmd := pipe.SMembers(c.ctx, boardCmtsKey(boardId))

	if _, err := pipe.Exec(c.ctx); err != nil {
		slog.Error("Failed to fetch board metadata pipeline", "err", err, "boardId", boardId)
		return nil, false
	}

	var b Board
	if err := boardCmd.Scan(&b); err != nil {
		slog.Error("Failed to scan board", "err", err)
		return nil, false
	}
	if b.Id == "" {
		return nil, false // Board not found
	}

	colIds := colIdsCmd.Val()
	userIds := userIdsCmd.Val()
	msgIds := msgIdsCmd.Val()
	cmtIds := cmtIdsCmd.Val()

	pipe2 := c.client.Pipeline()

	colCmds := make([]*redis.MapStringStringCmd, len(colIds))
	for i, id := range colIds {
		colCmds[i] = pipe2.HGetAll(c.ctx, boardColKey(boardId, id))
	}

	userCmds := make([]*redis.MapStringStringCmd, len(userIds))
	for i, id := range userIds {
		userCmds[i] = pipe2.HGetAll(c.ctx, boardUserKey(boardId, id))
	}

	msgCmds := make([]*redis.MapStringStringCmd, len(msgIds))
	for i, id := range msgIds {
		msgCmds[i] = pipe2.HGetAll(c.ctx, msgKey(id))
	}

	cmtCmds := make([]*redis.MapStringStringCmd, len(cmtIds))
	for i, id := range cmtIds {
		cmtCmds[i] = pipe2.HGetAll(c.ctx, msgKey(id))
	}

	if _, err := pipe2.Exec(c.ctx); err != nil {
		slog.Error("Failed to fetch board details pipeline", "err", err, "boardId", boardId)
		// We might choose to return partial data or fail. Here we fail safe.
		return nil, false
	}

	data := &BoardAggregatedData{
		Board:    &b,
		Columns:  make([]*BoardColumn, 0, len(colIds)),
		Users:    make([]*User, 0, len(userIds)),
		Messages: make([]*Message, 0, len(msgIds)),
		Comments: make([]*Message, 0, len(cmtIds)),
	}

	for _, cmd := range colCmds {
		var c BoardColumn
		if err := cmd.Scan(&c); err == nil {
			data.Columns = append(data.Columns, &c)
		}
	}
	for _, cmd := range userCmds {
		var u User
		if err := cmd.Scan(&u); err == nil {
			data.Users = append(data.Users, &u)
		}
	}
	for _, cmd := range msgCmds {
		var m Message
		if err := cmd.Scan(&m); err == nil {
			data.Messages = append(data.Messages, &m)
		}
	}
	for _, cmd := range cmtCmds {
		var m Message
		if err := cmd.Scan(&m); err == nil {
			data.Comments = append(data.Comments, &m)
		}
	}

	return data, true
}

func (c *RedisConnector) Close() {
	c.subscriber.Close()
	c.client.Close()
}
