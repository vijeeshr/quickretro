package main

// Store structure - Redis
/*
(KEY)board:{boardId}				(VALUE)board					Board - Redis Hash. The board details. Owned by single user.
(KEY)msg:{messageId}				(VALUE)message					Message - Redis Hash. Useful for fetch/add/update for an individual message.
(KEY)board:msg:{boardId}			(VALUE)[messageIds] 			Board-wise Messages - Redis Set. Useful for fetching list of messages.
(KEY)board:cmts:{boardId}      		(VALUE)[commentIds]      		Board-wise Comments - Redis Set. For fetching all comments.
(KEY)msg:likes:{messageId}			(VALUE)[userIds]				Likes - Redis Set. For recording likes/votes for a message.
(KEY)board:user:{boardId}:{userId}	(VALUE)User						User - Redis Hash. User master. Keeping as board specific.
(KEY)board:users:{boardId}			(VALUE)[userIds]				Board-wise Users - Redis Set. Useful for fetching members of a board.
(KEY)board:col:{boardId}:{colId}	(VALUE)column					Column - Redis Hash. Column definition for a Board.
(KEY)board:col:{boardId}			(Value)[colIds]					Board-wise columns - Redis Set. Just a list of colIds for a board.
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
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
		slog.Error("Cannot parsing Redis connection string", "details", err.Error())
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
		slog.Error("Cannot connect to Redis", "details", err.Error())
		os.Exit(1)
	}

	// return rdb.(*redis.Client), rdb.(*redis.Client).Subscribe(ctx)
	return &RedisConnector{client: rdb.(*redis.Client), subscriber: rdb.(*redis.Client).Subscribe(ctx), timeToLive: timeToLive, ctx: ctx}
}

func (c *RedisConnector) Subscribe(redisChannel ...string) {
	if err := c.subscriber.Subscribe(c.ctx, redisChannel...); err != nil {
		slog.Error("Unable to subscribe", "details", err.Error(), "channels", redisChannel)
	}
}

func (c *RedisConnector) Unsubscribe(redisChannel ...string) {
	if err := c.subscriber.Unsubscribe(c.ctx, redisChannel...); err != nil {
		slog.Error("Unable to Unsubscribe", "details", err.Error(), "channels", redisChannel)
	}
}

func (c *RedisConnector) Publish(redisChannel string, payload interface{}) {
	payload, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Marshal error when Publishing", "details", err.Error(), "channels", redisChannel, "payload", payload)
	}
	if err := c.client.Publish(c.ctx, redisChannel, payload).Err(); err != nil {
		slog.Error("Error when Publishing", "details", err.Error(), "channels", redisChannel, "payload", payload)
	}
}

func (c *RedisConnector) CreateBoard(b *Board, cols []*BoardColumn) bool {
	key := fmt.Sprintf("board:%s", b.Id)
	boardColsKey := fmt.Sprintf("board:col:%s", b.Id) // Boardwise-ColIds

	currentTime := time.Now().UTC()
	currentTimeUtcSeconds := currentTime.Unix()
	autoDeleteTime := currentTime.Add(c.timeToLive)
	autoDeleteTimeUtcSeconds := autoDeleteTime.Unix()

	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(c.ctx, key, "id", b.Id)
		pipe.HSet(c.ctx, key, "name", b.Name)
		pipe.HSet(c.ctx, key, "team", b.Team)
		pipe.HSet(c.ctx, key, "owner", b.Owner)
		pipe.HSet(c.ctx, key, "status", int(b.Status))
		pipe.HSet(c.ctx, key, "mask", b.Mask)
		pipe.HSet(c.ctx, key, "lock", b.Lock)
		pipe.HSet(c.ctx, key, "createdAtUtc", currentTimeUtcSeconds)
		pipe.HSet(c.ctx, key, "autoDeleteAtUtc", autoDeleteTimeUtcSeconds)
		// Columns
		for _, col := range cols {
			colKey := fmt.Sprintf("board:col:%s:%s", b.Id, col.Id)
			pipe.HSet(c.ctx, colKey, "id", col.Id)
			pipe.HSet(c.ctx, colKey, "text", col.Text)
			pipe.HSet(c.ctx, colKey, "isDefault", col.IsDefault)
			pipe.HSet(c.ctx, colKey, "color", col.Color)
			pipe.HSet(c.ctx, colKey, "pos", col.Position)
			// pipe.Expire(c.ctx, colKey, 2*time.Hour)
			pipe.ExpireAt(c.ctx, colKey, autoDeleteTime)
			pipe.SAdd(c.ctx, boardColsKey, col.Id)
		}
		pipe.ExpireAt(c.ctx, boardColsKey, autoDeleteTime)
		pipe.ExpireAt(c.ctx, key, autoDeleteTime)
		return nil
	})

	if err != nil {
		slog.Error("Failed to create board in Redis", "details", err.Error(), "board", b, "cols", cols)
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
	boardColsKey := fmt.Sprintf("board:col:%s", b.Id)
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
				colKey := fmt.Sprintf("board:col:%s:%s", b.Id, oldId)
				pipe.Del(c.ctx, colKey)
			}
		}

		// UPSERT (create or deep-update) columns
		for _, newCol := range newCols {

			oldCol, existed := oldMap[newCol.Id]
			colKey := fmt.Sprintf("board:col:%s:%s", b.Id, newCol.Id)

			// If didn’t exist → full create
			if !existed {
				pipe.HSet(c.ctx, colKey, "id", newCol.Id)
				pipe.HSet(c.ctx, colKey, "text", newCol.Text)
				pipe.HSet(c.ctx, colKey, "isDefault", newCol.IsDefault)
				pipe.HSet(c.ctx, colKey, "color", newCol.Color)
				pipe.HSet(c.ctx, colKey, "pos", newCol.Position)
				pipe.ExpireAt(c.ctx, colKey, autoDeleteTime)
				continue
			}

			// Deep diff: Only update changed fields
			if oldCol.Text != newCol.Text {
				pipe.HSet(c.ctx, colKey, "text", newCol.Text)
			}
			if oldCol.Color != newCol.Color {
				pipe.HSet(c.ctx, colKey, "color", newCol.Color)
			}
			if oldCol.IsDefault != newCol.IsDefault {
				pipe.HSet(c.ctx, colKey, "isDefault", newCol.IsDefault)
			}
			if oldCol.Position != newCol.Position {
				pipe.HSet(c.ctx, colKey, "pos", newCol.Position)
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
		slog.Error("Failed to deep diff reset columns", "details", err, "board", b.Id)
		return false
	}

	return true
}

func (c *RedisConnector) UpdateMasking(b *Board, mask bool) bool {
	// Todo: Deduplicate with UpdateBoardLock() & UpdateTimer()
	key := fmt.Sprintf("board:%s", b.Id)
	if _, err := c.client.HSet(c.ctx, key, "mask", mask).Result(); err != nil {
		slog.Error("Failed to mask/unmask", "details", err.Error(), "board", b)
		return false
	}
	return true
}

func (c *RedisConnector) UpdateBoardLock(b *Board, lock bool) bool {
	// Todo: Deduplicate with UpdateMasking() & UpdateTimer()
	key := fmt.Sprintf("board:%s", b.Id)
	if _, err := c.client.HSet(c.ctx, key, "lock", lock).Result(); err != nil {
		slog.Error("Failed to lock/unlock", "details", err.Error(), "board", b)
		return false
	}
	return true
}

func (c *RedisConnector) UpdateTimer(b *Board, expiryDurationInSeconds uint16) bool {
	// Todo: Deduplicate with UpdateMasking() & UpdateBoardLock()
	key := fmt.Sprintf("board:%s", b.Id)
	duration := time.Duration(expiryDurationInSeconds) * time.Second
	expiryTime := time.Now().UTC().Add(duration).Unix()

	if _, err := c.client.HSet(c.ctx, key, "timerExpiresAtUtc", expiryTime).Result(); err != nil {
		slog.Error("Failed to update board timer", "details", err.Error(), "board", b)
		return false
	}
	return true
}

func (c *RedisConnector) StopTimer(b *Board) bool {
	key := fmt.Sprintf("board:%s", b.Id)
	// To stop the timer, just reset the timerExpiresAtUtc to one second before current time.
	// This will cause the expiryTimeInSeconds to be sent as 0 (if expiryTime - curentTime is negative, its also sent as zero)
	// ...e.g check TimerEvent.broadcast, RegEvent.broadcast
	expiryTime := time.Now().UTC().Unix() - 1

	if _, err := c.client.HSet(c.ctx, key, "timerExpiresAtUtc", expiryTime).Result(); err != nil {
		slog.Error("Failed to update board timer during a 'Stop'", "details", err.Error(), "board", b)
		return false
	}
	return true
}

func (c *RedisConnector) BoardExists(boardId string) bool {
	key := fmt.Sprintf("board:%s", boardId)

	k, err := c.client.Exists(c.ctx, key).Result()
	if err != nil {
		slog.Error("Cannot find board in Redis", "details", err.Error(), "boardId", boardId)
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
	key := fmt.Sprintf("board:%s", boardId)

	if err := c.client.HGetAll(c.ctx, key).Scan(&b); err != nil {
		slog.Error("Failed to get board from Redis", "details", err.Error(), "boardId", boardId)
		return nil, false
	}
	// Assuming Id as empty to decide the key doesn't exist. This is done to avoid an additional EXISTS call to Redis.
	if b.Id == "" {
		return nil, false
	}

	return &b, true
}

func (c *RedisConnector) IsBoardOwner(boardId string, userId string) bool {
	key := fmt.Sprintf("board:%s", boardId)

	if userId == "" || boardId == "" {
		return false
	}

	owner, err := c.client.HGet(c.ctx, key, "owner").Result()
	if err != nil {
		slog.Error("Cannot find board in Redis", "details", err.Error(), "boardId", boardId)
		return false
	}

	return userId == owner
}

func (c *RedisConnector) IsBoardLocked(boardId string) bool {
	if boardId == "" {
		return true
	}

	key := fmt.Sprintf("board:%s", boardId)
	isLocked, err := c.client.HGet(c.ctx, key, "lock").Result()
	if err != nil {
		slog.Error("Cannot find board in Redis", "details", err.Error(), "boardId", boardId)
		return true
	}

	return isLocked == "1"
}

func (c *RedisConnector) GetBoardColumns(boardId string) ([]*BoardColumn, bool) {
	cols := make([]*BoardColumn, 0)

	key := fmt.Sprintf("board:col:%s", boardId)
	colIds, err := c.client.SMembers(c.ctx, key).Result()
	if err != nil {
		slog.Error("Failed to get columns from Redis", "details", err.Error(), "boardId", boardId)
		return cols, false
	}

	cmds, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		for _, id := range colIds {
			key := fmt.Sprintf("board:col:%s:%s", boardId, id)
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
			slog.Error("Failed to get/map column definition from Redis", "details", err.Error(), "boardId", boardId)
			continue
		}
		cols = append(cols, &c)
	}

	return cols, true
}

func (c *RedisConnector) CommitUserPresence(boardId string, user *User, isPresent bool) bool {
	userKey := fmt.Sprintf("board:user:%s:%s", boardId, user.Id)
	boardUsersKey := fmt.Sprintf("board:users:%s", boardId)

	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		if isPresent {
			pipe.HSet(c.ctx, userKey, "id", user.Id)
			pipe.HSet(c.ctx, userKey, "xid", user.Xid)
			pipe.HSet(c.ctx, userKey, "nickname", user.Nickname)
			pipe.SAdd(c.ctx, boardUsersKey, user.Id)
			pipe.Expire(c.ctx, userKey, c.timeToLive)       // Todo: We can try to expire this earlier by looking at Board.AutoDeleteAtUtc. But the requires a call to get board details. Skipping it for now.
			pipe.Expire(c.ctx, boardUsersKey, c.timeToLive) // Todo: We can try to expire this earlier by looking at Board.AutoDeleteAtUtc. But the requires a call to get board details. Skipping it for now.
			return nil
		} else {
			pipe.Del(c.ctx, userKey)
			pipe.SRem(c.ctx, boardUsersKey, user.Id)
			return nil
		}
	})

	if err != nil {
		slog.Error("Failed committing user presence to Redis", "details", err.Error(), "boardId", boardId, "user", user, "isPresent", isPresent)
		return false
	}

	return true
}

func (c *RedisConnector) RemoveUserPresence(boardId string, userId string) bool {
	userKey := fmt.Sprintf("board:user:%s:%s", boardId, userId)
	boardUsersKey := fmt.Sprintf("board:users:%s", boardId)

	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(c.ctx, userKey)
		pipe.SRem(c.ctx, boardUsersKey, userId)
		return nil
	})

	if err != nil {
		slog.Error("Failed removing user presence from Redis", "details", err.Error(), "boardId", boardId, "userId", userId)
		return false
	}

	return true
}

func (c *RedisConnector) GetUsersPresence(boardId string) ([]*User, bool) {
	users := make([]*User, 0)

	key := fmt.Sprintf("board:users:%s", boardId)
	userIds, err := c.client.SMembers(c.ctx, key).Result()
	if err != nil {
		slog.Error("Failed getting userIds from Redis", "details", err.Error(), "boardId", boardId)
		return users, false
	}

	cmds, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		for _, id := range userIds {
			key := fmt.Sprintf("board:user:%s:%s", boardId, id)
			pipe.HGetAll(c.ctx, key)
		}
		return nil
	})
	if err != nil {
		slog.Error("Error in GetUsersPresence", "details", err.Error())
		return users, false
	}

	for _, cmd := range cmds {
		var u User
		if err := cmd.(*redis.MapStringStringCmd).Scan(&u); err != nil {
			slog.Error("Failed getting/mapping users from Redis", "details", err.Error(), "boardId", boardId)
			continue
		}
		users = append(users, &u)
	}

	return users, true
}

func (c *RedisConnector) GetPresentUserIds(boardId string) ([]string, bool) {
	key := fmt.Sprintf("board:users:%s", boardId)
	ids, err := c.client.SMembers(c.ctx, key).Result()
	if err != nil {
		slog.Error("Failed getting userIds from Redis", "details", err.Error(), "boardId", boardId)
		return ids, false
	}
	return ids, true
}

func (c *RedisConnector) GetMessage(msgId string) (*Message, bool) {
	var message Message
	key := fmt.Sprintf("msg:%s", msgId)

	if err := c.client.HGetAll(c.ctx, key).Scan(&message); err != nil {
		slog.Error("Failed getting/mapping message from Redis", "details", err.Error(), "msgId", msgId)
		return nil, false
	}
	// Assuming Id as empty to decide the key doesn't exist. This is done to avoid an additional EXISTS call to Redis.
	if message.Id == "" {
		return nil, false
	}

	return &message, true
}

func (c *RedisConnector) GetMessages(boardId string) ([]*Message, bool) {
	key := fmt.Sprintf("board:msg:%s", boardId)
	messageIds, err := c.client.SMembers(c.ctx, key).Result()
	if err != nil {
		slog.Error("Failed getting messageIds from Redis", "details", err.Error(), "boardId", boardId)
		return make([]*Message, 0), false
	}
	return c.GetMessagesByIds(messageIds, boardId)
}

func (c *RedisConnector) GetComments(boardId string) ([]*Message, bool) {
	key := fmt.Sprintf("board:cmts:%s", boardId)
	commentIds, err := c.client.SMembers(c.ctx, key).Result()
	if err != nil {
		slog.Error("Failed getting commentIds from Redis", "details", err.Error(), "boardId", boardId)
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
			key := fmt.Sprintf("msg:%s", id)
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
			slog.Error("Failed getting/mapping message from Redis", "details", err.Error(), "boardId", boardId)
			continue
		}
		messages = append(messages, &m)
	}

	return messages, true
}

func (c *RedisConnector) GetLikesCount(msgId string) int64 {
	key := fmt.Sprintf("msg:likes:%s", msgId)

	count, err := c.client.SCard(c.ctx, key).Result()
	if err != nil {
		slog.Error("Failed getting likes count from Redis", "details", err.Error(), "msgId", msgId)
		return 0
	}
	return count
}

func (c *RedisConnector) HasLiked(msgId string, by string) bool {
	key := fmt.Sprintf("msg:likes:%s", msgId)
	if liked, err := c.client.SIsMember(c.ctx, key, by).Result(); err != nil {
		return false
	} else {
		return liked
	}
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
			key := fmt.Sprintf("msg:likes:%s", id)
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
	key := fmt.Sprintf("msg:%s", msg.Id)
	messagesKey := fmt.Sprintf("board:msg:%s", msg.Group)
	commentsKey := fmt.Sprintf("board:cmts:%s", msg.Group)

	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		// Always save the message/comment to the Hash
		pipe.HSet(c.ctx, key, "id", msg.Id)
		pipe.HSet(c.ctx, key, "by", msg.By)
		pipe.HSet(c.ctx, key, "nickname", msg.ByNickname)
		pipe.HSet(c.ctx, key, "group", msg.Group)
		pipe.HSet(c.ctx, key, "content", msg.Content)
		pipe.HSet(c.ctx, key, "category", msg.Category)
		pipe.HSet(c.ctx, key, "anon", msg.Anonymous)
		pipe.HSet(c.ctx, key, "pid", msg.ParentId)
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
		slog.Error("Failed to completely save message/comment to Redis", "details", err.Error(), "payloads", msg, "modes", modes)
		return false
	}

	return true
}

func (c *RedisConnector) Like(msgId string, by string, like bool) bool {
	var affected int64
	var err error
	key := fmt.Sprintf("msg:likes:%s", msgId)

	if like {
		affected, err = c.client.SAdd(c.ctx, key, by).Result() // Todo: Pipeline ?
		c.client.Expire(c.ctx, key, c.timeToLive)              // Todo: We can try to expire this earlier by looking at Board.AutoDeleteAtUtc. But the requires a call to get board details. Skipping it for now.
	} else {
		affected, err = c.client.SRem(c.ctx, key, by).Result()
	}
	if err != nil {
		slog.Error("Error when liking/unliking", "details", err.Error(), "msgId", msgId, "by", by, "like", like)
		return false
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
	key := fmt.Sprintf("msg:%s", msgId)
	likesKey := fmt.Sprintf("msg:likes:%s", msgId)
	messagesKey := fmt.Sprintf("board:msg:%s", group)
	commentsKey := fmt.Sprintf("board:cmts:%s", group)

	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		// Delete the top-level message and other related data
		pipe.Del(c.ctx, key, likesKey)
		pipe.SRem(c.ctx, messagesKey, msgId)
		for _, cid := range commentIds {
			cKey := fmt.Sprintf("msg:%s", cid)
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
	key := fmt.Sprintf("msg:%s", commentId)
	commentsKey := fmt.Sprintf("board:cmts:%s", group)

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
		(KEY)board:users:{boardId}				(VALUE)[userIds]				Board-wise Users - Redis Set. Useful for fetching members of a board.
		(KEY)board:user:{boardId}:{userId}	    (VALUE)User						User - Redis Hash. User master. Keeping as board specific.

		Columns
		(KEY)board:col:{boardId}				(Value)[colIds]					Board-wise columns - Redis Set. Just a list of colIds for a board.
		(KEY)board:col:{boardId}:{colId}		(VALUE)column					Column - Redis Hash. Column definition for a Board.
	*/
	ctx := c.ctx

	boardMsgsKey := fmt.Sprintf("board:msg:%s", boardId)
	boardCommsKey := fmt.Sprintf("board:cmts:%s", boardId)
	boardUsersKey := fmt.Sprintf("board:users:%s", boardId)
	boardColsKey := fmt.Sprintf("board:col:%s", boardId)
	boardKey := fmt.Sprintf("board:%s", boardId)

	// Collect all message Ids, comment Ids, user Ids, and column Ids
	// Pipeline SMEMBERS (read phase)
	readPipe := c.client.Pipeline()
	msgsCmd := readPipe.SMembers(ctx, boardMsgsKey)
	cmtsCmd := readPipe.SMembers(ctx, boardCommsKey)
	usrsCmd := readPipe.SMembers(ctx, boardUsersKey)
	colsCmd := readPipe.SMembers(ctx, boardColsKey)

	if _, err := readPipe.Exec(ctx); err != nil {
		slog.Error("Redis SMEMBERS pipeline failed in DeleteAll", "boardId", boardId, "error", err)
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
			likesKey := fmt.Sprintf("msg:likes:%s", msgId)
			msgKey := fmt.Sprintf("msg:%s", msgId)
			pipe.Del(ctx, likesKey, msgKey)
		}
		pipe.Del(ctx, boardMsgsKey)

		// Delete comments
		for _, cid := range commentIds {
			commentKey := fmt.Sprintf("msg:%s", cid)
			pipe.Del(ctx, commentKey)
		}
		pipe.Del(ctx, boardCommsKey)

		// Delete users
		for _, userId := range userIds {
			userKey := fmt.Sprintf("board:user:%s:%s", boardId, userId)
			pipe.Del(ctx, userKey)
		}
		pipe.Del(ctx, boardUsersKey)

		// Delete columns
		for _, colId := range colIds {
			colKey := fmt.Sprintf("board:col:%s:%s", boardId, colId)
			pipe.Del(ctx, colKey)
		}
		pipe.Del(ctx, boardColsKey)

		// Delete board hash
		pipe.Del(ctx, boardKey)

		return nil
	})

	if err != nil {
		slog.Error("Error when deleting all board data from Redis", "boardId", boardId, "details", err.Error())
		return false
	}

	return true
}

func (c *RedisConnector) UpdateCategory(category string, msgId string, commentIds []string) bool {
	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		// Update main message
		key := fmt.Sprintf("msg:%s", msgId)
		pipe.HSet(c.ctx, key, "category", category)
		// Update associated comments (if any)
		for _, cid := range commentIds {
			ckey := fmt.Sprintf("msg:%s", cid)
			pipe.HSet(c.ctx, ckey, "category", category)
		}
		return nil
	})

	if err != nil {
		slog.Error("Failed to update category", "error", err.Error(), "category", category, "msgId", msgId, "commentIds", commentIds)
		return false
	}
	return true
}

func (c *RedisConnector) Close() {
	c.subscriber.Close()
	c.client.Close()
}
