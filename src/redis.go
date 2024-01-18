package main

// Store structure - Redis
/*
(KEY)board:{boardId} 				(VALUE)board					Board - Redis Hash. The board details. Owned by single user.
(KEY)msg:{messageId}				(VALUE)message					Message - Redis Hash. Useful for fetch/add/update for an individual message.
(KEY)board:msg:{boardId} 			(VALUE)[messageIds] 			Board-wise Messages - Redis Set. Useful for fetching list of messages.
(KEY)msg:likes:{messageId} 			(VALUE)[userIds]				Likes - Redis Set. For recording likes/votes for a message

(KEY)board:user:{boardId}:{userId}	(VALUE)User						Users - Redis Hash. User master. Keeping as board specific.
(KEY)board:users:{boardId}			(VALUE)[userIds]				Board-wise Users - Redis Set. Useful for fetching members of a board.
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

type RedisConnector struct {
	client     *redis.Client
	subscriber *redis.PubSub
	ctx        context.Context
}

func NewRedisConnector(ctx context.Context) *RedisConnector {

	// rdb := redis.NewClient(&redis.Options{
	// 	Addr:     "localhost:6379",
	// 	Password: "", // no password set
	// 	DB:       0,  // use default DB
	// })

	// Get Redis server address from environment variable, defaulting to ":6379" for accessing redis from host.
	redisAddr := getEnv("REDIS_HOST", ":6379")

	// Todo: Add auth and pull from config
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		// Addrs:    []string{":6379"},
		// Addrs:    []string{"my-redis:6379"},
		Addrs:    []string{redisAddr},
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Error connecting to Redis:", err)
	}

	// return rdb.(*redis.Client), rdb.(*redis.Client).Subscribe(ctx)
	return &RedisConnector{client: rdb.(*redis.Client), subscriber: rdb.(*redis.Client).Subscribe(ctx), ctx: ctx}
}

// Todo: What about the context. Check.
func (c *RedisConnector) Subscribe(redisChannel ...string) {
	if err := c.subscriber.Subscribe(c.ctx, redisChannel...); err != nil {
		log.Fatal(err) // Todo: This will bring down app. Just do a log but don't exit. Check.
	}
}

// Todo: What about the context. Check.
func (c *RedisConnector) Unsubscribe(redisChannel ...string) {
	if err := c.subscriber.Unsubscribe(c.ctx, redisChannel...); err != nil {
		log.Printf("Error unsubscribing: %v", err)
	}
}

// Todo: What about the context. Check.
func (c *RedisConnector) Publish(redisChannel string, payload interface{}) {
	payload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Marshal error when Publishing. %v", err)
	}
	if err := c.client.Publish(c.ctx, redisChannel, payload).Err(); err != nil {
		log.Printf("Error when Publishing. %v", err)
	}
}

func (c *RedisConnector) CreateBoard(b *Board) bool {
	key := fmt.Sprintf("board:%s", b.Id)

	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		// Todo: SET TTL?
		pipe.HSet(c.ctx, key, "id", b.Id)
		pipe.HSet(c.ctx, key, "name", b.Name)
		pipe.HSet(c.ctx, key, "team", b.Team)
		pipe.HSet(c.ctx, key, "owner", b.Owner)
		pipe.HSet(c.ctx, key, "status", int(b.Status))
		pipe.HSet(c.ctx, key, "mask", b.Mask)
		pipe.HSet(c.ctx, key, "createdAtUtc", b.CreatedAtUtc)
		return nil
	})

	if err != nil {
		log.Printf("Error when saving to redis %v", err)
		return false
	}

	return true
}

func (c *RedisConnector) UpdateMasking(b *Board, mask bool) bool {
	key := fmt.Sprintf("board:%s", b.Id)
	if _, err := c.client.HSet(c.ctx, key, "mask", mask).Result(); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (c *RedisConnector) BoardExists(boardId string) bool {
	key := fmt.Sprintf("board:%s", boardId)

	k, err := c.client.Exists(c.ctx, key).Result()
	if err != nil {
		log.Printf("Error fetching from redis %v", err)
		return false
	}
	if k != 1 {
		log.Println("Non existent board")
		return false
	}

	return true
}

func (c *RedisConnector) GetBoard(boardId string) (*Board, bool) {
	var b Board
	key := fmt.Sprintf("board:%s", boardId)

	if err := c.client.HGetAll(c.ctx, key).Scan(&b); err != nil {
		log.Printf("Error fetching from redis %v", err)
		return nil, false
	}
	// Assuming Id as empty to decide the key doesn't exist. This is done to avoid an additional EXISTS call to Redis.
	if b.Id == "" {
		return nil, false
	}

	return &b, true
}

func (c *RedisConnector) CommitUserPresence(boardId string, user *User, isPresent bool) bool {
	userKey := fmt.Sprintf("board:user:%s:%s", boardId, user.Id)
	boardUsersKey := fmt.Sprintf("board:users:%s", boardId)

	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		// Todo: SET TTL?
		if isPresent {
			pipe.HSet(c.ctx, userKey, "id", user.Id)
			pipe.HSet(c.ctx, userKey, "xid", user.Xid)
			pipe.HSet(c.ctx, userKey, "nickname", user.Nickname)
			pipe.SAdd(c.ctx, boardUsersKey, user.Id)
			return nil
		} else {
			pipe.Del(c.ctx, userKey)
			pipe.SRem(c.ctx, boardUsersKey, user.Id)
			return nil
		}
	})

	if err != nil {
		log.Printf("Error when saving to redis %v", err)
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
		log.Printf("Error when saving to redis %v", err)
		return false
	}

	return true
}

func (c *RedisConnector) GetUsersPresence(boardId string) ([]*User, bool) {
	users := make([]*User, 0)

	key := fmt.Sprintf("board:users:%s", boardId)
	userIds, err := c.client.SMembers(c.ctx, key).Result()
	if err != nil {
		log.Println(err)
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
		return users, false
	}

	for _, cmd := range cmds {
		var u User
		if err := cmd.(*redis.MapStringStringCmd).Scan(&u); err != nil {
			log.Println(err)
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
		log.Println(err)
		return ids, false
	}
	return ids, true
}

func (c *RedisConnector) GetMessage(msgId string) (*Message, bool) {
	var message Message
	key := fmt.Sprintf("msg:%s", msgId)

	if err := c.client.HGetAll(c.ctx, key).Scan(&message); err != nil {
		log.Printf("Error fetching from redis %v", err)
		return nil, false
	}
	// Assuming Id as empty to decide the key doesn't exist. This is done to avoid an additional EXISTS call to Redis.
	if message.Id == "" {
		return nil, false
	}

	return &message, true
}

func (c *RedisConnector) GetMessages(boardId string) ([]*Message, bool) {
	messages := make([]*Message, 0)

	key := fmt.Sprintf("board:msg:%s", boardId)
	messageIds, err := c.client.SMembers(c.ctx, key).Result()
	if err != nil {
		log.Println(err)
		return messages, false
	}

	cmds, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		for _, id := range messageIds {
			key = fmt.Sprintf("msg:%s", id)
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
			log.Println(err)
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
		log.Println(err)
		return 0
	}
	return count
}

func (c *RedisConnector) GetLikesCountMultiple(msgIds ...string) map[string]int64 {
	result := make(map[string]int64)
	cmds, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		for _, id := range msgIds {
			key := fmt.Sprintf("msg:likes:%s", id)
			pipe.SCard(c.ctx, key)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
		return result
	}
	// Trusting Redis pipeline execution order for populating the result
	for i, cmd := range cmds {
		result[msgIds[i]] = cmd.(*redis.IntCmd).Val()
	}
	return result
}

func (c *RedisConnector) HasLiked(msgId string, by string) bool {
	key := fmt.Sprintf("msg:likes:%s", msgId)
	if liked, err := c.client.SIsMember(c.ctx, key, by).Result(); err != nil {
		return false
	} else {
		return liked
	}
}

// Todo: Unused. Just added for checking.
// Store helper - DTO
type LikedBy struct {
	By    string
	Liked bool
}

// Todo: Unused. Just added for checking.
func (c *RedisConnector) HasLikedList(msgId string, by ...string) []*LikedBy {
	var likes []*LikedBy
	key := fmt.Sprintf("msg:likes:%s", msgId)
	cmds, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		for _, v := range by {
			pipe.SIsMember(c.ctx, key, v)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
		return nil
	}
	// Trusting Redis pipeline execution order for populating the result
	for i, cmd := range cmds {
		likes = append(likes, &LikedBy{By: by[i], Liked: cmd.(*redis.BoolCmd).Val()})
	}
	return likes
}

func (c *RedisConnector) Save(msg *Message) bool {
	msgKey := fmt.Sprintf("msg:%s", msg.Id)
	boardKey := fmt.Sprintf("board:msg:%s", msg.Group)

	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		// Todo: SET TTL?
		pipe.HSet(c.ctx, msgKey, "id", msg.Id)
		pipe.HSet(c.ctx, msgKey, "by", msg.By)
		pipe.HSet(c.ctx, msgKey, "nickname", msg.ByNickname)
		pipe.HSet(c.ctx, msgKey, "group", msg.Group)
		pipe.HSet(c.ctx, msgKey, "content", msg.Content)
		pipe.HSet(c.ctx, msgKey, "category", msg.Category)
		pipe.SAdd(c.ctx, boardKey, msg.Id)
		return nil
	})

	if err != nil {
		log.Printf("Error when saving to redis %v", err)
		return false
	}

	return true
}

func (c *RedisConnector) Like(msgId string, by string, like bool) bool {
	var affected int64
	var err error
	key := fmt.Sprintf("msg:likes:%s", msgId)

	if like {
		affected, err = c.client.SAdd(c.ctx, key, by).Result()
	} else {
		affected, err = c.client.SRem(c.ctx, key, by).Result()
	}
	if err != nil {
		log.Println(err)
		return false
	}
	if affected == 0 {
		if like {
			log.Println("Cannot like a message which is already liked")
		} else {
			log.Println("Message must be liked for it to be unliked")
		}
		return false
	}
	return true
}

func (c *RedisConnector) DeleteMessage(msg *Message) bool {
	msgKey := fmt.Sprintf("msg:%s", msg.Id)
	likesKey := fmt.Sprintf("msg:likes:%s", msg.Id)
	boardKey := fmt.Sprintf("board:msg:%s", msg.Group)

	_, err := c.client.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(c.ctx, msgKey, likesKey)
		pipe.SRem(c.ctx, boardKey, msg.Id)
		return nil
	})

	// Todo: Should individual results be checked from the pipeline response?
	if err != nil {
		log.Printf("Error when deleting from redis %v", err)
		return false
	}
	return true
}

func (c *RedisConnector) Close() {
	c.subscriber.Close()
	c.client.Close()
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
