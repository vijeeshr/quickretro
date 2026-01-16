package scenarios

// cd tests/functional
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

	t.Run("Registration: Alice registers", func(t *testing.T) {
		require.NoError(t, userA.Register())

		wantReg := harness.RegisterResponse{
			Type:         "reg",
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

	// Todo: Server validates if same msgId that is attached to a board is not "accidently" updated by another user from another board. Add test for that.

	t.Run("Message: Alice sends new message", func(t *testing.T) {
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

	t.Run("Message: Alice edits message", func(t *testing.T) {
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

	/*
		// 3. Register Users
		err := userA.Register()
		require.NoError(t, err)
		regRespUserA, err := userA.WaitForEvent("reg", 2*time.Second) // A receives reg response
		require.NoError(t, err)
		_, err = userB.WaitForEvent("joining", 2*time.Second) // B receives joining response
		require.NoError(t, err)

		var r harness.RegisterResponse
		err = json.Unmarshal(regRespUserA.Raw, &r)
		require.NoError(t, err)
		require.Equal(t, true, r.IsBoardOwner) // A should be board owner

		userA.FlushEvents()
		userB.FlushEvents()

		err = userB.Register()
		require.NoError(t, err)
		regRespUserB, err := userB.WaitForEvent("reg", 2*time.Second) // B receives reg response
		require.NoError(t, err)

		err = json.Unmarshal(regRespUserB.Raw, &r)
		require.NoError(t, err)
		require.Equal(t, false, r.IsBoardOwner) // B should NOT be board owner

		userA.FlushEvents()
		userB.FlushEvents()
	*/

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
