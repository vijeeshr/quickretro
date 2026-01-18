package scenarios

// cd tests/e2e
// go test -v ./scenarios/...
// go test -v ./scenarios/locking_test.go ./scenarios/messaging_test.go

import (
	"bytes"
	"crypto/tls"
	"e2e_tests/harness"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const BaseURL = "http://localhost:8080" // https://localhost

func CreateBoard(t *testing.T, ownerId string) string {
	payload := map[string]any{
		"name":  "E2E Test Board",
		"team":  "Test Team",
		"owner": ownerId,
		// "cfTurnstileResponse": "1x00000000000000000000AA", // Dummy always pass token
		"cfTurnstileResponse": "",
		"columns": []map[string]any{
			{"id": "col01", "text": "What went well", "isDefault": true, "color": "green", "pos": 1},
			{"id": "col02", "text": "Challenges", "isDefault": true, "color": "red", "pos": 2},
			{"id": "col03", "text": "Action Items", "isDefault": true, "color": "yellow", "pos": 3},
			{"id": "col04", "text": "Appreciations", "isDefault": true, "color": "fuchsia", "pos": 4},
			{"id": "col05", "text": "Improvements", "isDefault": true, "color": "orange", "pos": 5},
		},
	}

	body, _ := json.Marshal(payload)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         "localhost",
		},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Post(BaseURL+"/api/board/create", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	return result["id"]
}

func TestMessagingFlow(t *testing.T) {
	// Shared Setup
	userAId := "user-a"
	userBId := "user-b"

	// Create Board
	boardId := CreateBoard(t, userAId)
	t.Logf("Created board: %s", boardId)

	userA := harness.NewUser(userAId, "Alice", boardId)
	userB := harness.NewUser(userBId, "Bob", boardId)

	// Connect Users
	require.NoError(t, userA.Connect(BaseURL))
	defer userA.Close()
	require.NoError(t, userB.Connect(BaseURL))
	defer userB.Close()

	// TODO: Add test to check connection failure to non existent board

	t.Run("Registration: Alice registers", func(t *testing.T) {
		require.NoError(t, userA.Register())

		wantReg := harness.RegisterResponse{
			Type:         "reg",
			BoardMasking: true,
			BoardLock:    false,
			IsBoardOwner: true,
		}
		t.Run("Alice receives own reg event, shown as board owner", func(t *testing.T) {
			var got harness.RegisterResponse
			userA.MustWaitForEvent(t, "reg", &got)
			require.Equal(t, wantReg, got)
		})

		wantJoining := harness.UserJoiningResponse{
			Type:     "joining",
			Nickname: userA.Nickname,
			Xid:      "xid-" + userA.Id,
		}
		t.Run("Bob receives joining event for Alice", func(t *testing.T) {
			var got harness.UserJoiningResponse
			userB.MustWaitForEvent(t, "joining", &got)
			require.Equal(t, wantJoining, got)
		})

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Registration: Bob registers", func(t *testing.T) {
		require.NoError(t, userB.Register())

		t.Run("Bob receives own reg event", func(t *testing.T) {
			var regResp harness.RegisterResponse
			userB.MustWaitForEvent(t, "reg", &regResp)
			require.False(t, regResp.IsBoardOwner) // Bob should NOT be board owner
		})

		t.Run("Alice receives joining event for Bob", func(t *testing.T) {
			var joinResp harness.UserJoiningResponse
			userA.MustWaitForEvent(t, "joining", &joinResp)
			require.Equal(t, "Bob", joinResp.Nickname)
		})

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Registration: Do NOT allow registration to non-existent board", func(t *testing.T) {
		userA.Board = "nonexistent-board"
		userB.Board = "nonexistent-board"
		require.NoError(t, userB.Register())

		require.NoError(t, userB.MustNotReceiveAnyEvent())
		require.NoError(t, userA.MustNotReceiveAnyEvent())

		// reset objects back to original board
		userA.Board = boardId
		userB.Board = boardId
	})

	t.Run("Message: Create", func(t *testing.T) {
		content := "Hello from Alice"
		category := "col01"
		msgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)

		want := harness.MessageResponse{
			Type:       "msg",
			Id:         msgId,
			ParentId:   "",
			ByXid:      "xid-" + userA.Id,
			ByNickname: userA.Nickname,
			Content:    content,
			Category:   category,
			Likes:      int64(0),
			Liked:      false,
			Mine:       false,
			Anonymous:  false,
		}

		// Send message
		require.NoError(t, userA.SendMessage(msgId, content, category))

		t.Run("Bob receives it", func(t *testing.T) {
			var got harness.MessageResponse
			userB.MustWaitForEvent(t, "msg", &got)
			require.Equal(t, want, got)
		})

		t.Run("Alice(sender) too receives it", func(t *testing.T) {
			var got harness.MessageResponse
			userA.MustWaitForEvent(t, "msg", &got)
			want.Mine = true // "Mine" will be true for the message owner
			require.Equal(t, want, got)
		})

		// Todo: Delete this message for idempotency

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Message: Update", func(t *testing.T) {
		msgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)

		content := "Second message"
		category := "col01"

		want := harness.MessageResponse{
			Type:       "msg",
			Id:         msgId,
			ParentId:   "",
			ByXid:      "xid-" + userA.Id,
			ByNickname: userA.Nickname,
			Content:    content,
			Category:   category,
			Likes:      int64(0),
			Liked:      false,
			Mine:       false,
			Anonymous:  false,
		}

		// Alice sends new message
		require.NoError(t, userA.SendMessage(msgId, content, category))
		userA.FlushEvents()
		userB.FlushEvents()

		t.Run("Alice edits previous message content, received by all", func(t *testing.T) {
			updatedContent := "Updated second message"
			want.Content = updatedContent
			// Edit message
			require.NoError(t, userA.SendMessage(msgId, updatedContent, category))
			// Bob receives edited message
			var got harness.MessageResponse
			userB.MustWaitForEvent(t, "msg", &got)
			require.Equal(t, want, got)

			userA.FlushEvents()
			userB.FlushEvents()
		})

		t.Run("Alice tampers existing message edit process with category change", func(t *testing.T) {
			updatedCategory := "col02"
			want.Content = content
			// Edit same message again, but also try updating category.
			// Catergory in response should be same as old one. Only "content" must be updated.
			require.NoError(t, userA.SendMessage(msgId, content, updatedCategory))

			t.Run("Category should not be changed in response", func(t *testing.T) {
				var got harness.MessageResponse
				userB.MustWaitForEvent(t, "msg", &got)
				// log.Printf("After: %+v", msgResp)
				require.Equal(t, want, got)
			})

			userA.FlushEvents()
			userB.FlushEvents()
		})

		// Todo: Delete this message for idempotency?
	})

	t.Run("Message: Delete", func(t *testing.T) {
		content := "This message should be deleted"
		category := "col03"

		t.Run("User deletes own message", func(t *testing.T) {
			msgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)
			want := harness.DeleteMessageResponse{
				Type: "del",
				Id:   msgId,
			}
			// Create message
			require.NoError(t, userB.SendMessage(msgId, content, category))
			userA.FlushEvents()
			userB.FlushEvents()
			// Message creator deletes it
			require.NoError(t, userB.DeleteMessage(msgId))

			var got harness.DeleteMessageResponse
			userA.MustWaitForEvent(t, "del", &got)
			require.Equal(t, want, got)
			userB.MustWaitForEvent(t, "del", &got)
			require.Equal(t, want, got)
		})

		t.Run("Board owner deletes another user's message", func(t *testing.T) {
			msgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)
			// Create message
			require.NoError(t, userB.SendMessage(msgId, content, category))
			userA.FlushEvents()
			userB.FlushEvents()
			// Board owner deletes it
			require.NoError(t, userA.DeleteMessage(msgId))

			var got harness.DeleteMessageResponse
			userA.MustWaitForEvent(t, "del", &got)
			require.Equal(t, msgId, got.Id)
			userB.MustWaitForEvent(t, "del", &got)
			require.Equal(t, msgId, got.Id)
		})

		t.Run("Guest user should NOT delete another user's message", func(t *testing.T) {
			content = "Unauthorizd user attempted to delete this card, but failed :)"
			msgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)
			// Create message
			require.NoError(t, userA.SendMessage(msgId, content, category))
			userA.FlushEvents()
			userB.FlushEvents()
			// Another user attempts to delete it
			require.NoError(t, userB.DeleteMessage(msgId))

			require.NoError(t, userA.MustNotReceiveEvent("del"))
			require.NoError(t, userB.MustNotReceiveEvent("del"))
			// require.NoError(t, userA.MustNotReceiveAnyEvent())
			// require.NoError(t, userB.MustNotReceiveAnyEvent())
		})
	})

	t.Run("Message: Anonymous", func(t *testing.T) {
		msgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)
		category := "col02"

		t.Run("New anonymous message is sent and received", func(t *testing.T) {
			anonymous := true
			emptyXid := ""
			emptyNickname := ""
			content := "Guess who am I?"
			// Send
			require.NoError(t, userA.SendAnonymousMessage(msgId, content, category, emptyXid, emptyNickname, anonymous))
			// Bob receives it
			var got harness.MessageResponse
			userB.MustWaitForEvent(t, "msg", &got)
			require.Empty(t, got.ByXid)
			require.Empty(t, got.ByNickname)
			require.True(t, got.Anonymous)
			require.Equal(t, content, got.Content)

			userA.FlushEvents()
			userB.FlushEvents()
		})

		t.Run("Existing anonymous message must NOT be made non-anonymous later", func(t *testing.T) {
			// Try updating xid, nickname, anonymous alongwith the content
			anonymous := false
			xid := "xid-" + userA.Id
			nickname := userA.Nickname
			updatedContent := "Guess who am I? Guess who again!!"
			// Send
			require.NoError(t, userA.SendAnonymousMessage(msgId, updatedContent, category, xid, nickname, anonymous))
			// Bob receives it
			var got harness.MessageResponse
			userB.MustWaitForEvent(t, "msg", &got)
			// xid, nickname, anonymous fields shouldn't be changed
			require.Empty(t, got.ByXid)
			require.Empty(t, got.ByNickname)
			require.True(t, got.Anonymous)
			require.Equal(t, updatedContent, got.Content) // content should be updated

			userA.FlushEvents()
			userB.FlushEvents()
		})

		t.Run("xid and nickname must be empty in response, even if passed with values in event request", func(t *testing.T) {
			id := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)
			anonymous := true
			xid := "xid-" + userA.Id
			nickname := userA.Nickname
			content := "Oops! Passed Xid and Nickname"
			// Send
			require.NoError(t, userA.SendAnonymousMessage(id, content, category, xid, nickname, anonymous))
			// Bob receives it
			var got harness.MessageResponse
			userB.MustWaitForEvent(t, "msg", &got)
			require.Empty(t, got.ByXid)
			require.Empty(t, got.ByNickname)
			require.True(t, got.Anonymous)
			require.Equal(t, content, got.Content)

			userA.FlushEvents()
			userB.FlushEvents()
		})

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Mask:", func(t *testing.T) {

		wantMasked := harness.MaskResponse{
			Type: "mask",
			Mask: true,
		}
		wantRevealed := harness.MaskResponse{
			Type: "mask",
			Mask: false,
		}

		t.Run("Owner can Mask or Reveal messages", func(t *testing.T) {
			var got harness.MaskResponse

			// Owner reveals board
			require.NoError(t, userA.Mask(false))
			userB.MustWaitForEvent(t, "mask", &got)
			require.Equal(t, wantRevealed, got)

			// Attempt to reveal/Unmask an already revealed board is silently discarded
			require.NoError(t, userA.Mask(false))
			userB.MustNotReceiveEvent("mask")

			// Owner masks board
			require.NoError(t, userA.Mask(true))
			userB.MustWaitForEvent(t, "mask", &got)
			require.Equal(t, wantMasked, got)

			// Attempt to mask an already masked board is silently discarded
			require.NoError(t, userA.Mask(true))
			userB.MustNotReceiveEvent("mask")

			// Remove mask for further tests downstream
			require.NoError(t, userA.Mask(false))

			userA.FlushEvents()
			userB.FlushEvents()
		})

		t.Run("Guest users cannot issue mask event", func(t *testing.T) {
			require.NoError(t, userB.Mask(true))
			userA.MustNotReceiveEvent("mask")
		})

	})

	t.Run("Lock:", func(t *testing.T) {

		wantLocked := harness.LockResponse{
			Type: "lock",
			Lock: true,
		}
		wantUnlocked := harness.LockResponse{
			Type: "lock",
			Lock: false,
		}

		t.Run("Owner can Lock or Unlock", func(t *testing.T) {
			var got harness.LockResponse

			// Owner locks board
			require.NoError(t, userA.LockBoard(true))
			userB.MustWaitForEvent(t, "lock", &got)
			require.Equal(t, wantLocked, got)

			// Attempt to lock an already locked board is silently discarded
			require.NoError(t, userA.LockBoard(true))
			userB.MustNotReceiveEvent("lock")

			// Owner unlocks board
			require.NoError(t, userA.LockBoard(false))
			userB.MustWaitForEvent(t, "lock", &got)
			require.Equal(t, wantUnlocked, got)

			// Attempt to unlock an already unlocked board is silently discarded
			require.NoError(t, userA.LockBoard(false))
			userB.MustNotReceiveEvent("lock")

			userA.FlushEvents()
			userB.FlushEvents()
		})

		t.Run("Guest users cannot issue lock event", func(t *testing.T) {
			require.NoError(t, userB.LockBoard(true))
			userA.MustNotReceiveEvent("lock")
		})

	})

	// Todo: Add tests

	// Connection: User attempts to connect to non-existant board should fail
	// Registration: Check count of messages/comments when user register later on in a board
	// Message: Updating message in a different board should fail. (Server validates if same msgId that is attached to a board is not "accidently" updated by another user from another board. Add test for that.)
	// Message: Sending message in non-existant board should fail
	// Message: Updating message of another user should fail (Even for board owner)
	// Message: Guest user deleting another user's message/comment should fail
	// Message: Deleting a message, should delete associated comments. (Todo: How to figure that a comment is present, may be a "reg" response's GetBoardAggregrate should retun all comments. Start by not passing the associated commentIds[] in the main message's del request. Is this test valid? Orphaned comments are auto-deleted. How does the system process their presence.)
	// Message: Creating/Editing/Deleting in a locked board should fail
	// Comment: Create/Edit/Delete comment
	// Comment: Board owner can delete another users comment
	// Comment: Update/Deleting another user's comment should fail
	// Comment: Send comment to a non-existant message should fail
	// Comment: Associating comment of message1 to message2 should fail
	// Comment: Sending/Deleting comment to locked board should fail
	// CategoryChange: User can move message to another category. Any associated comment category must also be updated.
	// CategoryChange: Owner can move any user's message
	// CategoryChange: Moving message to non-existant/invalid or disabled category should fail
	// CategoryChange: Moving category of another message in another board should fail
	// CategoryChange: Moving category in a locked board should fail
	// Like: Check like toggle
	// Like: Liking and already likes message should fail
	// Like: Toggling likes in a locked board should fail
	// Timer: Only board owner can Start/Stop timer
	// Timer: Guest user cannot Start/Stop timer
	// ColumnEditing:
	// BoardDeletion:
	// UserLeaving:

	// Check that A sees B joining
	// Note: Due to concurrency, A may see B joining before or after A's reg response depending on timing,
	// but WaitForEvent logic in harness just pulls next.
	// Ideally we should wait for "joining" event for userA
	// For simplicity, let's assume robust matching or just check that flow proceeds.

	// Flush initial events for cleanliness if needed, or just let them sit in channel.
	// userA.FlushEvents()
	// userB.FlushEvents()

	// // ################################################################################
	// // 4. User A sends message
	// msgId, err := userA.SendMessage("Hello from A", "col1")
	// require.NoError(t, err)

	// // 5. User B receives message
	// // B might have other events queued (like A joining if B connected early, or B's own reg).
	// // We loop to find the message
	// var receivedMsg *harness.MessageResponse

	// // We expect a 'msg' event
	// timeout := time.After(5 * time.Second)
	// found := false
	// for {
	// 	select {
	// 	case ev := <-userB.Events:
	// 		if ev.Type == "msg" {
	// 			// verify it matches
	// 			var m harness.MessageResponse
	// 			err := json.Unmarshal(ev.Raw, &m)
	// 			t.Logf("Received message event. ID: %s (Expected: %s), Content: %s", m.Id, msgId, m.Content)
	// 			if err == nil && m.Id == msgId {
	// 				receivedMsg = &m
	// 				found = true
	// 			}
	// 		}
	// 	case <-timeout:
	// 		t.Fatal("Timeout waiting for message")
	// 	}
	// 	if found {
	// 		break
	// 	}
	// }

	// assert.Equal(t, "Hello from A", receivedMsg.Content)
	// assert.Equal(t, "col1", receivedMsg.Category)

	// // 6. User B likes message
	// err = userB.LikeMessage(msgId, true)
	// require.NoError(t, err)

	// // 7. User A receives like update
	// // A receives "like" event
	// found = false
	// timeout = time.After(5 * time.Second)
	// for {
	// 	select {
	// 	case ev := <-userA.Events:
	// 		if ev.Type == "like" {
	// 			// verify it matches
	// 			// var l harness.LikeMessageEvent
	// 			// Note: The broadcast payload for 'like' is actually a MessageResponse with updated like count/status?
	// 			// Checking eventtypes.go:
	// 			// func (i *LikeMessageEvent) Broadcast(m *Message, h *Hub) { base := m.NewLikeResponse() ... }
	// 			// NewLikeResponse returns Type: "like", Id: ..., Likes: ..., Liked: ...

	// 			// Let's decode as generic map to be safe or specific struct if known
	// 			// user.go defines LikeMessageEvent which matches the Handle payload, but Broadcast payload might be different.
	// 			// In eventtypes.go:
	// 			/*
	// 			   type LikeResponse struct {
	// 			       Type  string `json:"typ"`
	// 			       Id    string `json:"id"`
	// 			       Likes int    `json:"likes"`
	// 			       Liked bool   `json:"liked"`
	// 			   }
	// 			*/
	// 			// Wait, I didn't define LikeResponse in events.go. I should check logic.
	// 			// Assuming payload has 'id', 'likes'.

	// 			var payload map[string]interface{}
	// 			json.Unmarshal(ev.Raw, &payload)
	// 			if id, ok := payload["id"].(string); ok && id == msgId {
	// 				if likes, ok := payload["likes"].(float64); ok && likes >= 1 {
	// 					found = true
	// 				}
	// 			}
	// 		}
	// 	case <-timeout:
	// 		t.Fatal("Timeout waiting for like update")
	// 	}
	// 	if found {
	// 		break
	// 	}
	// }
	// // ################################################################################
}
