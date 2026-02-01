package scenarios

import (
	"e2e_tests/harness"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Tests registration flow explicitly
func TestRegistration(t *testing.T) {
	const (
		MaskedDefault = true
		LockedDefault = false
	)
	_, userA, userB := harness.SetupTest(t, false)

	// Registration: Alice registers
	t.Run("Alice registers", func(t *testing.T) {
		require.NoError(t, userA.Register())

		t.Run("Alice receives own reg event, shown as board owner", func(t *testing.T) {
			var got harness.RegisterResponse
			userA.MustWaitForEvent(t, "reg", &got)
			require.True(t, got.IsBoardOwner)
			require.Equal(t, 1, len(got.Users))
			require.Equal(t, harness.UserDetails{
				Nickname: userA.Nickname,
				Xid:      "xid-" + userA.Id,
			}, got.Users[0])
			require.Equal(t, 0, len(got.Messages))
			require.Equal(t, 0, len(got.Comments))
			require.ElementsMatch(t, []*harness.BoardColumn{
				{Id: "col01", Text: "What went well", IsDefault: true, Color: "green", Position: 1},
				{Id: "col02", Text: "Challenges", IsDefault: true, Color: "red", Position: 2},
				{Id: "col03", Text: "Action Items", IsDefault: true, Color: "yellow", Position: 3},
				{Id: "col04", Text: "Appreciations", IsDefault: true, Color: "fuchsia", Position: 4},
				{Id: "col05", Text: "Improvements", IsDefault: true, Color: "orange", Position: 5},
			}, got.BoardColumns)
			require.Equal(t, LockedDefault, got.BoardLock)
			require.Equal(t, MaskedDefault, got.BoardMasking)
		})

		t.Run("Bob receives joining event for Alice", func(t *testing.T) {
			var got harness.UserJoiningResponse
			userB.MustWaitForEvent(t, "joining", &got)
			require.Equal(t, "xid-"+userA.Id, got.Xid)
			require.Equal(t, userA.Nickname, got.Nickname)
		})

		userA.FlushEvents()
		userB.FlushEvents()
	})

	// Registration: Bob registers
	t.Run("Bob registers", func(t *testing.T) {
		require.NoError(t, userB.Register())

		t.Run("Bob receives own reg event", func(t *testing.T) {
			var got harness.RegisterResponse
			userB.MustWaitForEvent(t, "reg", &got)
			require.False(t, got.IsBoardOwner) // Bob should NOT be board owner
			require.Equal(t, 2, len(got.Users))
			require.ElementsMatch(t, []harness.UserDetails{
				{Nickname: userA.Nickname, Xid: "xid-" + userA.Id},
				{Nickname: userB.Nickname, Xid: "xid-" + userB.Id},
			}, got.Users)
		})

		t.Run("Alice receives joining event for Bob", func(t *testing.T) {
			var got harness.UserJoiningResponse
			userA.MustWaitForEvent(t, "joining", &got)
			require.Equal(t, "xid-"+userB.Id, got.Xid)
			require.Equal(t, userB.Nickname, got.Nickname)
		})

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Alice registers again", func(t *testing.T) {
		// Add a message and a comment
		category := "col01"
		// Add a message
		msgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userB.Id)
		require.NoError(t, userB.SendMessage(msgId, "First message", category))
		// Add a comment to the message
		cmtId := fmt.Sprintf("cmt-%d-%s", time.Now().UnixNano(), userB.Id)
		require.NoError(t, userB.SendComment(cmtId, "First comment", category, msgId))
		// Clear responses for above events
		userA.FlushEvents()
		userB.FlushEvents()

		// Register again
		require.NoError(t, userA.Register())

		t.Run("Reg event response should have messages and comment", func(t *testing.T) {
			var got harness.RegisterResponse
			userA.MustWaitForEvent(t, "reg", &got)

			require.Equal(t, 2, len(got.Users))
			require.ElementsMatch(t, []harness.UserDetails{
				{Nickname: userA.Nickname, Xid: "xid-" + userA.Id},
				{Nickname: userB.Nickname, Xid: "xid-" + userB.Id},
			}, got.Users)

			require.Equal(t, 1, len(got.Messages))
			require.Equal(t, msgId, got.Messages[0].Id)

			require.Equal(t, 1, len(got.Comments))
			require.Equal(t, cmtId, got.Comments[0].Id)
		})

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Reg event should have correct mask and lock status", func(t *testing.T) {
		// Register
		require.NoError(t, userA.Register())
		// Check default values in Register response
		var got harness.RegisterResponse
		userA.MustWaitForEvent(t, "reg", &got)
		require.Equal(t, MaskedDefault, got.BoardMasking)
		require.Equal(t, LockedDefault, got.BoardLock)

		// Unmask and lock
		require.NoError(t, userA.Mask(!MaskedDefault))
		require.NoError(t, userA.LockBoard(!LockedDefault))
		// Ignore mask and lock responses
		userA.FlushEvents()
		userB.FlushEvents()

		// Register again
		require.NoError(t, userA.Register())
		// Register response should show new values
		userA.MustWaitForEvent(t, "reg", &got)
		require.Equal(t, !MaskedDefault, got.BoardMasking)
		require.Equal(t, !LockedDefault, got.BoardLock)

		// Restore to defaults: Mask and Unlock for further tests
		require.NoError(t, userA.Mask(MaskedDefault))
		require.NoError(t, userA.LockBoard(LockedDefault))

		userA.FlushEvents()
		userB.FlushEvents()
	})

	// TODO: Check after deleting board
	// // Negative Test: Registration to non-existent board
	// t.Run("Do NOT allow registration to non-existent board", func(t *testing.T) {
	// 	originalBoard := userA.Board
	// 	userA.Board = "nonexistent-board"

	// 	require.NoError(t, userA.Register())

	// 	require.NoError(t, userA.MustNotReceiveAnyEvent())

	// 	userA.Board = originalBoard // Restore
	// })
}

func TestMessageLifecycle(t *testing.T) {
	_, userA, userB := harness.SetupTest(t, true)

	content := "First message"
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

	t.Run("New message received by all", func(t *testing.T) {
		// Send message
		require.NoError(t, userA.SendMessage(msgId, content, category))

		// Another user receives it
		var got harness.MessageResponse
		userB.MustWaitForEvent(t, "msg", &got)
		require.Equal(t, want, got)

		// Sender too receives it
		userA.MustWaitForEvent(t, "msg", &got)
		require.True(t, got.Mine)
	})

	// TODO: Check after deleting board
	// t.Run("New message to non-existent board should fail", func(t *testing.T) {
	// 	originalBoard := userA.Board
	// 	userA.Board = "nonexistent-board"

	// 	require.NoError(t, userA.SendMessage(msgId, content, category))
	// 	require.NoError(t, userA.MustNotReceiveAnyEvent())

	// 	userA.Board = originalBoard

	// 	userA.FlushEvents()
	// 	userB.FlushEvents()
	// })

	t.Run("Edit existing message", func(t *testing.T) {
		content = "First message is edited"
		require.NoError(t, userA.SendMessage(msgId, content, category))

		var got harness.MessageResponse
		userB.MustWaitForEvent(t, "msg", &got)
		require.Equal(t, content, got.Content)

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("User cannot edit another user's message", func(t *testing.T) {
		bobsMsgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userB.Id)
		bobsContent := "Bob's first message"
		require.NoError(t, userB.SendMessage(bobsMsgId, bobsContent, category))
		userA.FlushEvents()
		userB.FlushEvents()

		// Alice attempts to update Bob's message
		require.NoError(t, userA.SendMessage(bobsMsgId, "Hacking Bob", category))

		require.NoError(t, userB.MustNotReceiveAnyEvent())
		require.NoError(t, userA.MustNotReceiveAnyEvent())
	})

	t.Run("Update: Category change ignored during edit if not allowed?", func(t *testing.T) {
		// The original test said: "Sender tampers existing message edit process with category change"
		// "Category in response should be same as old one. Only 'content' must be updated."
		updatedCategory := "col02"
		content = "First message is edited, again!"

		require.NoError(t, userA.SendMessage(msgId, content, updatedCategory))

		var got harness.MessageResponse
		userB.MustWaitForEvent(t, "msg", &got)
		require.Equal(t, content, got.Content)
		require.Equal(t, category, got.Category) // Should NOT correspond to updatedCategory

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Delete: User deletes own message", func(t *testing.T) {
		delMsgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userB.Id)
		require.NoError(t, userB.SendMessage(delMsgId, "To be deleted", category))
		userA.FlushEvents()
		userB.FlushEvents()

		require.NoError(t, userB.DeleteMessage(delMsgId))

		var got harness.DeleteMessageResponse
		userA.MustWaitForEvent(t, "del", &got)
		require.Equal(t, delMsgId, got.Id)

		userB.MustWaitForEvent(t, "del", &got)
		require.Equal(t, delMsgId, got.Id)

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Delete: Board owner deletes another user's message", func(t *testing.T) {
		delMsgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userB.Id)
		require.NoError(t, userB.SendMessage(delMsgId, "To be deleted by Admin", category))
		userA.FlushEvents()
		userB.FlushEvents()

		require.NoError(t, userA.DeleteMessage(delMsgId))

		var got harness.DeleteMessageResponse
		userA.MustWaitForEvent(t, "del", &got)
		require.Equal(t, delMsgId, got.Id)

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Delete: Guest user should NOT delete another user's message", func(t *testing.T) {
		// Alice creates message
		delMsgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)
		require.NoError(t, userA.SendMessage(delMsgId, "Alice's Protected Msg", category))
		userA.FlushEvents()
		userB.FlushEvents()

		// Bob tries to delete it
		require.NoError(t, userB.DeleteMessage(delMsgId))

		require.NoError(t, userA.MustNotReceiveEvent("del"))
		require.NoError(t, userB.MustNotReceiveEvent("del"))

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Deleting a message should delete associated comments(if commentIds passed)", func(t *testing.T) {
		// Setup
		// Bob creates message, and associates a comment with it
		delMsgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userB.Id)
		delCmtId := fmt.Sprintf("cmt-%d-%s", time.Now().UnixNano(), userB.Id)
		require.NoError(t, userB.SendMessage(delMsgId, "Root Message", category))
		require.NoError(t, userB.SendComment(delCmtId, "Comment to root message", category, delMsgId))

		userA.FlushEvents()
		userB.FlushEvents()

		// Act
		// Bob deletes message with its comments
		// NOTE: With current design limitation, associated commentIds must be passed with the payload when deleting a message..
		// ..if not passed, those comments end up orphaned, and only removed after Redis TTL expiry
		require.NoError(t, userB.DeleteMessageWithComments(delMsgId, delCmtId))
		userA.FlushEvents()
		userB.FlushEvents()

		// Assert
		// Both message and comment must be deleted
		require.NoError(t, userB.Register())
		var regResp harness.RegisterResponse
		userB.MustWaitForEvent(t, "reg", &regResp)
		require.NotContains(t, extractIDs(regResp.Messages), delMsgId)
		require.NotContains(t, extractIDs(regResp.Comments), delCmtId)
	})

	t.Run("Deleting a message should NOT delete other non-associated comments", func(t *testing.T) {
		// Setup
		// Bob creates 2 messages with comments
		// message1->comment1,comment11
		rootMsg1Id := fmt.Sprintf("msg1-%d-%s", time.Now().UnixNano(), userB.Id)
		cmt1Id := fmt.Sprintf("cmt1-%d-%s", time.Now().UnixNano(), userB.Id)
		cmt11Id := fmt.Sprintf("cmt11-%d-%s", time.Now().UnixNano(), userB.Id)
		require.NoError(t, userB.SendMessage(rootMsg1Id, "Root Message1", category))
		require.NoError(t, userB.SendComment(cmt1Id, "Comment1 to root message1", category, rootMsg1Id))
		require.NoError(t, userB.SendComment(cmt11Id, "Comment11 to root message1", category, rootMsg1Id))
		// message2->comment2
		rootMsg2Id := fmt.Sprintf("msg2-%d-%s", time.Now().UnixNano(), userB.Id)
		cmt2Id := fmt.Sprintf("cmt2-%d-%s", time.Now().UnixNano(), userB.Id)
		require.NoError(t, userB.SendMessage(rootMsg2Id, "Root Message2", category))
		require.NoError(t, userB.SendComment(cmt2Id, "Comment2 to root message2", category, rootMsg2Id))

		userA.FlushEvents()
		userB.FlushEvents()

		// Act
		// Attempt to delete message1->comment1,comment2* (comment2 belong to message2)
		// Ideal delete payload: message1->comment1,comment11
		require.NoError(t, userB.DeleteMessageWithComments(rootMsg1Id, cmt1Id, cmt2Id))
		var delResp harness.DeleteMessageResponse
		userB.MustWaitForEvent(t, "del", &delResp)
		require.Equal(t, rootMsg1Id, delResp.Id)
		userA.FlushEvents()
		userB.FlushEvents()

		// Assert
		// Message1 must be deleted. Comment1(Orphaned), Message2, Comment2 must be present.
		require.NoError(t, userB.Register())
		var regResp harness.RegisterResponse
		userB.MustWaitForEvent(t, "reg", &regResp)

		require.NotContains(t, extractIDs(regResp.Messages), rootMsg1Id) // deleted message1
		require.NotContains(t, extractIDs(regResp.Comments), cmt1Id)     // deleted message1->comment1
		require.Contains(t, extractIDs(regResp.Messages), rootMsg2Id)    // untouched message2
		require.Contains(t, extractIDs(regResp.Comments), cmt2Id)        // untouched message2->comment2, event though (comment2) passed in request
		require.Contains(t, extractIDs(regResp.Comments), cmt11Id)       // orphaned (message1 does not exist, but comment11 does)
	})

	t.Run("Messages should NOT be added/updated/deleted in a locked board", func(t *testing.T) {
		// Setup
		const (
			lock   = true
			Unlock = false
		)

		// Add message. There will be an attempt to delete this after the board is locked.
		lockedBoardMsgId1 := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)
		require.NoError(t, userA.SendMessage(lockedBoardMsgId1, "There was an attempt to delete this message when the board was locked. It failed :)", category))
		// Lock board
		require.NoError(t, userA.LockBoard(lock))
		userA.FlushEvents()
		userB.FlushEvents()

		// Act
		// Attempt to add a message
		lockedBoardMsgId2 := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)
		require.NoError(t, userA.SendMessage(lockedBoardMsgId2, "Trying to add/update message to locked board", category))
		// Attempt to delete an existing message in a locked board
		require.NoError(t, userA.DeleteMessage(lockedBoardMsgId1))

		// Assert
		// No events must be received
		require.NoError(t, userA.MustNotReceiveAnyEvent())

		// Cleanup
		// Unlock board for further tests, if any
		require.NoError(t, userA.LockBoard(Unlock))

		userA.FlushEvents()
		userB.FlushEvents()
	})

	// // TODO: Check this. Currently the message is created in non-existent (or existent but inactive) category...
	// // ..Check for impact of adding another redis call to validate in a hot path.
	// t.Run("New Message to non-existent category should fail", func(t *testing.T) {
	// 	nonCatMsgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)

	// 	require.NoError(t, userA.SendMessage(nonCatMsgId, "Message content to non-existent category", "col08"))
	// 	require.NoError(t, userA.MustNotReceiveAnyEvent())

	// 	userA.FlushEvents()
	// 	userB.FlushEvents()
	// })
}

func extractIDs(msgs []harness.MessageResponse) []string {
	ids := make([]string, 0, len(msgs))
	for _, m := range msgs {
		ids = append(ids, m.Id)
	}
	return ids
}

func TestAnonymousMessaging(t *testing.T) {
	_, userA, userB := harness.SetupTest(t, true)

	msgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)
	category := "col02"

	t.Run("New anonymous message is sent and received", func(t *testing.T) {
		anonymous := true
		content := "Guess who am I?"
		require.NoError(t, userA.SendAnonymousMessage(msgId, content, category, "", anonymous))

		var got harness.MessageResponse
		userB.MustWaitForEvent(t, "msg", &got)
		require.Empty(t, got.ByXid)
		require.Empty(t, got.ByNickname)
		require.True(t, got.Anonymous)
		require.Equal(t, content, got.Content)

		userA.FlushEvents()
	})

	t.Run("Existing anonymous message must NOT be made non-anonymous later", func(t *testing.T) {
		anonymous := false // TRYING to change it to false
		nickname := userA.Nickname
		updatedContent := "Revealing myself?"

		require.NoError(t, userA.SendAnonymousMessage(msgId, updatedContent, category, nickname, anonymous))

		var got harness.MessageResponse
		userB.MustWaitForEvent(t, "msg", &got)
		// Should still be anonymous in response
		require.Empty(t, got.ByXid)
		require.Empty(t, got.ByNickname)
		require.True(t, got.Anonymous)
		require.Equal(t, updatedContent, got.Content)
	})
}

func TestCommenting(t *testing.T) {
	_, userA, userB := harness.SetupTest(t, true)

	category := "col02"
	rootMsgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userB.Id)
	require.NoError(t, userB.SendMessage(rootMsgId, "Bob's Root Message", category))
	userA.FlushEvents()
	userB.FlushEvents()

	t.Run("Create/Update/Delete comments", func(t *testing.T) {
		cmtId := fmt.Sprintf("cmt-%d-%s", time.Now().UnixNano(), userA.Id)

		// Create new comment
		cmtContent := "Alice's first comment"
		require.NoError(t, userA.SendComment(cmtId, cmtContent, category, rootMsgId))
		// Assert
		var got harness.MessageResponse
		userB.MustWaitForEvent(t, "msg", &got)
		require.Equal(t, cmtContent, got.Content)
		require.Equal(t, cmtId, got.Id)
		require.Equal(t, rootMsgId, got.ParentId)
		require.False(t, got.Mine)
		// Assert for sender too
		userA.MustWaitForEvent(t, "msg", &got)
		require.True(t, got.Mine)

		// Edit previous comment
		updatedCmdContent := "Alice edited first comment"
		require.NoError(t, userA.SendComment(cmtId, updatedCmdContent, category, rootMsgId))
		// Assert changes
		userB.MustWaitForEvent(t, "msg", &got)
		require.Equal(t, updatedCmdContent, got.Content)
		require.Equal(t, cmtId, got.Id)
		userA.FlushEvents()
		userB.FlushEvents()

		// Delete previous comment
		require.NoError(t, userA.DeleteComment(cmtId))
		var delResp harness.DeleteMessageResponse
		userB.MustWaitForEvent(t, "del", &delResp)
		require.Equal(t, cmtId, delResp.Id)

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Comment must NOT be associated to a non-existent message", func(t *testing.T) {
		cmtId := fmt.Sprintf("cmt-%d-%s", time.Now().UnixNano(), userA.Id)
		require.NoError(t, userA.SendComment(cmtId, "Dangling comment", category, "nonexistent"))
		require.NoError(t, userA.MustNotReceiveAnyEvent())
	})

	t.Run("Comment must NOT be associated to a comment (No nesting)", func(t *testing.T) {
		// Create a valid comment
		cmtId1 := fmt.Sprintf("cmt1-%d-%s", time.Now().UnixNano(), userA.Id)
		require.NoError(t, userA.SendComment(cmtId1, "Tesing for nested comment", category, rootMsgId))
		userA.FlushEvents()
		userB.FlushEvents()

		// Try to attach another comment to it
		cmtId2 := fmt.Sprintf("cmt2-%d-%s", time.Now().UnixNano(), userB.Id)
		require.NoError(t, userB.SendComment(cmtId2, "Nested comment", category, cmtId1))

		require.NoError(t, userB.MustNotReceiveAnyEvent())
	})

	t.Run("Board owner can delete another user's comment", func(t *testing.T) {
		// Guest use creates new comment
		cmtId := fmt.Sprintf("cmt-%d-%s", time.Now().UnixNano(), userB.Id)
		require.NoError(t, userB.SendComment(cmtId, "Bob's comment that will be deleted by board owner", category, rootMsgId))
		userA.FlushEvents()
		userB.FlushEvents()

		// Board owner deletes it
		require.NoError(t, userA.DeleteComment(cmtId))

		// Assert
		var got harness.DeleteMessageResponse
		userB.MustWaitForEvent(t, "del", &got)
		require.Equal(t, cmtId, got.Id)

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Guest user must NOT delete another user's comment", func(t *testing.T) {
		// User creates new comment
		cmtId := fmt.Sprintf("cmt-%d-%s", time.Now().UnixNano(), userA.Id)
		require.NoError(t, userA.SendComment(cmtId, "Alice's comment that Bob(guest user) tries to delete, but fails :)", category, rootMsgId))
		userA.FlushEvents()
		userB.FlushEvents()

		// Guest user tries to delete it
		require.NoError(t, userB.DeleteComment(cmtId))

		// Assert
		require.NoError(t, userA.MustNotReceiveAnyEvent())

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("User(even board owner) must NOT edit another user's comment", func(t *testing.T) {
		// Guest use creates new comment
		cmtId := fmt.Sprintf("cmt-%d-%s", time.Now().UnixNano(), userB.Id)
		require.NoError(t, userB.SendComment(cmtId, "Bob's comment that Alice(boad owner), tries to change, and fails :)", category, rootMsgId))
		userA.FlushEvents()
		userB.FlushEvents()

		// Board owner tries to edit it
		require.NoError(t, userA.SendComment(cmtId, "Hack attempt from Alice", category, rootMsgId))

		// Assert
		require.NoError(t, userA.MustNotReceiveAnyEvent())

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Comment's parent Message cannot be changed", func(t *testing.T) {
		// Setup
		// Bob creates 2 messages with comments
		// message1->comment1
		rootMsg1Id := fmt.Sprintf("msg1-%d-%s", time.Now().UnixNano(), userB.Id)
		cmt1Id := fmt.Sprintf("cmt1-%d-%s", time.Now().UnixNano(), userB.Id)
		require.NoError(t, userB.SendMessage(rootMsg1Id, "Root Message1", category))
		require.NoError(t, userB.SendComment(cmt1Id, "Comment1 to root message1", category, rootMsg1Id))
		// message2->comment2
		rootMsg2Id := fmt.Sprintf("msg2-%d-%s", time.Now().UnixNano(), userB.Id)
		cmt2Id := fmt.Sprintf("cmt2-%d-%s", time.Now().UnixNano(), userB.Id)
		require.NoError(t, userB.SendMessage(rootMsg2Id, "Root Message2", category))
		require.NoError(t, userB.SendComment(cmt2Id, "Comment2 to root message2", category, rootMsg2Id))

		userA.FlushEvents()
		userB.FlushEvents()

		// Act
		// Try to associate comment1 to message2
		require.NoError(t, userB.SendComment(cmt1Id, "Comment1 to root message1", category, rootMsg2Id))

		// Assert
		require.NoError(t, userB.MustNotReceiveAnyEvent())

		userA.FlushEvents()
		userB.FlushEvents()
	})

	// Locked board test case can be ignored. That's covered by "messages" test cases.
}

func TestBoardControl(t *testing.T) {
	_, userA, userB := harness.SetupTest(t, true)

	t.Run("Owner can Mask/Reveal", func(t *testing.T) {
		var got harness.MaskResponse

		// Board starts masked, so unmask first
		require.NoError(t, userA.Mask(false))
		userB.MustWaitForEvent(t, "mask", &got)
		require.False(t, got.Mask)

		// Then mask again
		require.NoError(t, userA.Mask(true))
		userB.MustWaitForEvent(t, "mask", &got)
		require.True(t, got.Mask)
	})

	t.Run("Guest users cannot mask", func(t *testing.T) {
		require.NoError(t, userB.Mask(true))
		require.NoError(t, userB.MustNotReceiveEvent("mask"))
	})

	t.Run("Owner can Lock/Unlock", func(t *testing.T) {
		var got harness.LockResponse
		require.NoError(t, userA.LockBoard(true))
		userB.MustWaitForEvent(t, "lock", &got)
		require.True(t, got.Lock)

		require.NoError(t, userA.LockBoard(false))
		userB.MustWaitForEvent(t, "lock", &got)
		require.False(t, got.Lock)
	})

	t.Run("Guest users cannot lock", func(t *testing.T) {
		require.NoError(t, userB.LockBoard(true))
		require.NoError(t, userB.MustNotReceiveEvent("lock"))
	})
}

func TestTimer(t *testing.T) {
	_, userA, userB := harness.SetupTest(t, true)

	t.Run("Owner can Start/Stop timer", func(t *testing.T) {
		timerDurationToExpirySeconds := uint16(3600)
		// Start
		require.NoError(t, userA.StartTimer(timerDurationToExpirySeconds))
		var got harness.TimerResponse
		userB.MustWaitForEvent(t, "timer", &got)
		require.Greater(t, got.ExpiresInSeconds, timerDurationToExpirySeconds-uint16(2)) // Flaky, or just use uint16(0)
		userA.FlushEvents()
		userB.FlushEvents()

		t.Run("Timer info is also part of Reg response", func(t *testing.T) {
			require.NoError(t, userB.Register())
			var got harness.RegisterResponse
			userB.MustWaitForEvent(t, "reg", &got)
			require.Greater(t, got.TimerExpiresInSeconds, timerDurationToExpirySeconds-uint16(2)) // Flaky, or just use uint16(0)
			userA.FlushEvents()
			userB.FlushEvents()
		})

		// Stop
		require.NoError(t, userA.StopTimer(uint16(0)))
		userB.MustWaitForEvent(t, "timer", &got)
		require.Equal(t, uint16(0), got.ExpiresInSeconds)
		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Timer should NOT start with invalid range (valid range 1 - 3600 seconds)", func(t *testing.T) {
		require.NoError(t, userA.StartTimer(3601))
		require.NoError(t, userB.MustNotReceiveAnyEvent())
		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Should NOT Stop non-running timer", func(t *testing.T) {
		// Check if Timer is not running currently
		require.NoError(t, userB.Register())
		var regRes harness.RegisterResponse
		userB.MustWaitForEvent(t, "reg", &regRes)
		require.Equal(t, uint16(0), regRes.TimerExpiresInSeconds)
		userA.FlushEvents()
		userB.FlushEvents()

		// Try to Stop a timer that isn't running
		require.NoError(t, userA.StopTimer(uint16(0)))
		// Assert
		require.NoError(t, userB.MustNotReceiveAnyEvent())
		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Should NOT Start running timer again", func(t *testing.T) {
		// Check if Timer is not running currently
		require.NoError(t, userB.Register())
		var regRes harness.RegisterResponse
		userB.MustWaitForEvent(t, "reg", &regRes)
		require.Equal(t, uint16(0), regRes.TimerExpiresInSeconds)
		userA.FlushEvents()
		userB.FlushEvents()

		// Start twice in succession
		timerDurationToExpirySeconds1 := uint16(10)
		require.NoError(t, userA.StartTimer(timerDurationToExpirySeconds1))
		userA.FlushEvents()
		userB.FlushEvents()
		timerDurationToExpirySeconds2 := uint16(100)
		require.NoError(t, userA.StartTimer(timerDurationToExpirySeconds2))

		// Assert
		require.NoError(t, userB.MustNotReceiveAnyEvent())

		// Stop timer for further tests
		require.NoError(t, userA.StopTimer(uint16(0)))

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Guest user cannot Start/Stop timer", func(t *testing.T) {
		// Check if Timer is not running currently
		require.NoError(t, userB.Register())
		var regRes harness.RegisterResponse
		userB.MustWaitForEvent(t, "reg", &regRes)
		require.Equal(t, uint16(0), regRes.TimerExpiresInSeconds)
		userA.FlushEvents()
		userB.FlushEvents()

		// Attempt to start Timer
		require.NoError(t, userB.StartTimer(30))
		require.NoError(t, userA.MustNotReceiveAnyEvent())

		userA.FlushEvents()
		userB.FlushEvents()
	})

	// Timer operations ignore board lock. So no tests for that.
}

func TestMessageCategoryChange(t *testing.T) {
	_, userA, userB := harness.SetupTest(t, true)

	t.Run("User can move own message and associated comments to another category", func(t *testing.T) {
		// Setup
		category := "col01"
		rootMsgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userB.Id)
		cmtId := fmt.Sprintf("cmt-%d-%s", time.Now().UnixNano(), userB.Id)
		require.NoError(t, userB.SendMessage(rootMsgId, "Root Message", category))
		require.NoError(t, userB.SendComment(cmtId, "Comment", category, rootMsgId))
		userA.FlushEvents()
		userB.FlushEvents()
		// Act
		newCategory := "col03"
		require.NoError(t, userB.ChangeCategoryOfMessageAndComments(rootMsgId, category, newCategory, cmtId))
		// Assert
		var got harness.CategoryChangeResponse
		userA.MustWaitForEvent(t, "catchng", &got)
		require.Equal(t, newCategory, got.NewCategory)
		require.Equal(t, rootMsgId, got.MessageId)
		userA.FlushEvents()
		userB.FlushEvents()

		t.Run("Comment category would match its parent Message category", func(t *testing.T) {
			// Act
			// Find category of the attached comment by firing an event
			require.NoError(t, userB.SendComment(cmtId, "Comment", category, rootMsgId))
			// Assert
			var got harness.MessageResponse
			userA.MustWaitForEvent(t, "msg", &got)
			require.Equal(t, newCategory, got.Category)

			userA.FlushEvents()
			userB.FlushEvents()
		})
	})

	t.Run("User can move own message to another category", func(t *testing.T) {
		// Setup
		category := "col01"
		rootMsgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userB.Id)
		cmtId := fmt.Sprintf("cmt-%d-%s", time.Now().UnixNano(), userB.Id)
		require.NoError(t, userB.SendMessage(rootMsgId, "Root Message", category))
		require.NoError(t, userB.SendComment(cmtId, "Comment", category, rootMsgId))
		userA.FlushEvents()
		userB.FlushEvents()
		// Act
		newCategory := "col02"
		require.NoError(t, userB.ChangeCategoryOfMessage(rootMsgId, category, newCategory))
		// Assert
		var got harness.CategoryChangeResponse
		userA.MustWaitForEvent(t, "catchng", &got)
		require.Equal(t, newCategory, got.NewCategory)
		require.Equal(t, rootMsgId, got.MessageId)
		userA.FlushEvents()
		userB.FlushEvents()

		t.Run("Comment category would NOT be updated to new one", func(t *testing.T) {
			// Act
			// Find category of the attached comment by firing an event
			require.NoError(t, userB.SendComment(cmtId, "Comment", category, rootMsgId))
			// Assert
			// This is expected due to a design limitation
			var got harness.MessageResponse
			userA.MustWaitForEvent(t, "msg", &got)
			require.Equal(t, category, got.Category) // Still points to old category

			userA.FlushEvents()
			userB.FlushEvents()
		})
	})

	t.Run("Board owner can change any message's category", func(t *testing.T) {
		// Setup
		// Guest user creates message
		category := "col01"
		rootMsgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userB.Id)
		require.NoError(t, userB.SendMessage(rootMsgId, "Root Message", category))
		userA.FlushEvents()
		userB.FlushEvents()
		// Act
		// Board owner (Alice) changes category
		newCategory := "col04"
		require.NoError(t, userA.ChangeCategoryOfMessage(rootMsgId, category, newCategory))
		// Assert
		var got harness.CategoryChangeResponse
		userB.MustWaitForEvent(t, "catchng", &got)
		require.Equal(t, newCategory, got.NewCategory)
		require.Equal(t, rootMsgId, got.MessageId)

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Guest user cannot change another user's message category", func(t *testing.T) {
		// Setup
		// Alice creates message
		category := "col01"
		rootMsgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)
		require.NoError(t, userA.SendMessage(rootMsgId, "Alice's Message", category))
		userA.FlushEvents()
		userB.FlushEvents()
		// Act
		// Guest user tries to change category
		newCategory := "col04"
		require.NoError(t, userB.ChangeCategoryOfMessage(rootMsgId, category, newCategory))
		// Assert
		require.NoError(t, userA.MustNotReceiveAnyEvent())

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Should NOT move message to disabled/inactive category", func(t *testing.T) {
		// Setup
		// Alice creates message in col01
		category := "col01"
		rootMsgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)
		require.NoError(t, userA.SendMessage(rootMsgId, "Alice's Message", category))
		userA.FlushEvents()
		userB.FlushEvents()
		// Disable col05
		newCols := []*harness.BoardColumn{
			{Id: "col01", Text: "What went well", Color: "green", Position: 1, IsDefault: true},
			{Id: "col02", Text: "Challenges", Color: "red", Position: 2, IsDefault: true},
			{Id: "col03", Text: "Action Items", Color: "yellow", Position: 3, IsDefault: true},
			{Id: "col04", Text: "Appreciations", Color: "fuchsia", Position: 4, IsDefault: true},
		}
		require.NoError(t, userA.ChangeColumns(newCols))
		var colChangeRes harness.ColumnsChangeResponse
		userB.MustWaitForEvent(t, "colreset", &colChangeRes)
		require.Equal(t, 4, len(colChangeRes.BoardColumns))
		userA.FlushEvents()
		userB.FlushEvents()

		// Act
		// Board owner (Alice) tries to move message to inactive col05
		disabledCategory := "col05"
		require.NoError(t, userA.ChangeCategoryOfMessage(rootMsgId, category, disabledCategory))

		// Assert
		require.NoError(t, userB.MustNotReceiveAnyEvent())

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Should NOT move message to non-existent category", func(t *testing.T) {
		// Setup
		// User creates message
		category := "col01"
		rootMsgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)
		require.NoError(t, userA.SendMessage(rootMsgId, "Alice's Message", category))
		userA.FlushEvents()
		userB.FlushEvents()
		// Act
		// Try to move to a category that doesn't exist
		nonExistentCategory := "col08"
		require.NoError(t, userA.ChangeCategoryOfMessage(rootMsgId, category, nonExistentCategory))
		// Assert
		require.NoError(t, userB.MustNotReceiveAnyEvent())

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Should NOT change category in a locked board", func(t *testing.T) {
		// Setup
		const (
			lock   = true
			Unlock = false
		)

		// Add message.
		category := "col01"
		lockedBoardMsgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)
		require.NoError(t, userA.SendMessage(lockedBoardMsgId, fmt.Sprintf("This message is in category:%s", category), category))
		// Lock board
		require.NoError(t, userA.LockBoard(lock))
		userA.FlushEvents()
		userB.FlushEvents()

		// Act
		newCategory := "col02"
		require.NoError(t, userA.ChangeCategoryOfMessage(lockedBoardMsgId, category, newCategory))

		// Assert
		require.NoError(t, userA.MustNotReceiveAnyEvent())

		// Cleanup
		// Unlock board for further tests, if any
		require.NoError(t, userA.LockBoard(Unlock))

		userA.FlushEvents()
		userB.FlushEvents()
	})

}

func TestLikes(t *testing.T) {
	_, userA, userB := harness.SetupTest(t, true)

	const (
		liked    = true
		notLiked = false
	)
	// Create a message to test likes
	category := "col01"
	rootMsgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userB.Id)
	require.NoError(t, userB.SendMessage(rootMsgId, "Message from Bob to test likes", category))
	userA.FlushEvents()
	userB.FlushEvents()

	t.Run("Toggle like/unlike", func(t *testing.T) {
		// Bob likes a message
		require.NoError(t, userB.LikeMessage(rootMsgId, liked))

		// Assert
		var likeRes harness.LikeMessageResponse
		userB.MustWaitForEvent(t, "like", &likeRes)
		require.Equal(t, rootMsgId, likeRes.Id)
		require.Equal(t, int64(1), likeRes.Likes)
		require.Equal(t, liked, likeRes.Liked)
		userA.MustWaitForEvent(t, "like", &likeRes)
		require.Equal(t, notLiked, likeRes.Liked)

		// Assert likes information in MessageResponse
		require.NoError(t, userB.SendMessage(rootMsgId, "Message from Bob to test likes. Edited just in case.", category))
		var msgRes harness.MessageResponse
		userB.MustWaitForEvent(t, "msg", &msgRes)
		require.Equal(t, int64(1), msgRes.Likes)
		require.Equal(t, liked, msgRes.Liked)
		userA.MustWaitForEvent(t, "msg", &msgRes)
		require.Equal(t, notLiked, msgRes.Liked)

		// Assert likes information in RegResponse
		require.NoError(t, userB.Register())
		var regRes harness.RegisterResponse
		userB.MustWaitForEvent(t, "reg", &regRes)
		require.Equal(t, int64(1), regRes.Messages[0].Likes)
		require.Equal(t, liked, regRes.Messages[0].Liked)
		userA.FlushEvents()
		userB.FlushEvents()
		require.NoError(t, userA.Register())
		userA.MustWaitForEvent(t, "reg", &regRes)
		require.Equal(t, notLiked, regRes.Messages[0].Liked)
		userA.FlushEvents()
		userB.FlushEvents()

		// Bob Unlikes a message
		require.NoError(t, userB.LikeMessage(rootMsgId, notLiked))
		// Assert
		userB.MustWaitForEvent(t, "like", &likeRes)
		require.Equal(t, int64(0), likeRes.Likes)
		require.Equal(t, notLiked, likeRes.Liked)

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Liking an already liked message(and vice-versa) is silently skipped", func(t *testing.T) {
		// Bob likes a message, and tries to like same message again
		require.NoError(t, userB.LikeMessage(rootMsgId, liked))
		userA.FlushEvents()
		userB.FlushEvents()
		require.NoError(t, userB.LikeMessage(rootMsgId, liked))
		// assert
		require.NoError(t, userB.MustNotReceiveAnyEvent())

		// Same for Unlike
		require.NoError(t, userB.LikeMessage(rootMsgId, notLiked))
		userA.FlushEvents()
		userB.FlushEvents()
		require.NoError(t, userB.LikeMessage(rootMsgId, notLiked))
		// assert
		require.NoError(t, userB.MustNotReceiveAnyEvent())

		userA.FlushEvents()
		userB.FlushEvents()
	})

	// Todo this needs to be fixed server side.
	// t.Run("Should NOT toggle likes in a locked board", func(t *testing.T) {
	// 	require.NoError(t, userA.LockBoard(true))
	// 	userA.FlushEvents()
	// 	userB.FlushEvents()

	// 	require.NoError(t, userB.LikeMessage(rootMsgId, liked))
	// 	require.NoError(t, userB.MustNotReceiveAnyEvent())

	// 	userA.FlushEvents()
	// 	userB.FlushEvents()
	// })
}

func TestColumnEditing(t *testing.T) {
	_, userA, userB := harness.SetupTest(t, true)

	t.Run("Owner can update columns", func(t *testing.T) {
		newCols := []*harness.BoardColumn{
			{Id: "col01", Text: "Start", Color: "green", Position: 1, IsDefault: false},
			{Id: "col02", Text: "Stop", Color: "red", Position: 2, IsDefault: false},
			{Id: "col03", Text: "Continue", Color: "yellow", Position: 3, IsDefault: false},
			{Id: "col04", Text: "Appreciations", Color: "fuchsia", Position: 4, IsDefault: false},
		}

		require.NoError(t, userA.ChangeColumns(newCols))

		var got harness.ColumnsChangeResponse
		userB.MustWaitForEvent(t, "colreset", &got)
		require.Equal(t, 4, len(got.BoardColumns))
		require.Equal(t, "Start", got.BoardColumns[0].Text)
		userA.FlushEvents()
		userB.FlushEvents()

		// Reg response should also show above columns
		require.NoError(t, userA.Register())

		var regRes harness.RegisterResponse
		userA.MustWaitForEvent(t, "reg", &regRes)
		require.Equal(t, 4, len(regRes.BoardColumns))

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Change column order, remove col03, add col04", func(t *testing.T) {
		newCols := []*harness.BoardColumn{
			{Id: "col04", Text: "Appreciations", Color: "fuchsia", Position: 1, IsDefault: true},
			{Id: "col02", Text: "Stop", Color: "red", Position: 2, IsDefault: false},
			{Id: "col01", Text: "Start", Color: "green", Position: 3, IsDefault: false},
		}

		require.NoError(t, userA.ChangeColumns(newCols))

		var got harness.ColumnsChangeResponse
		userB.MustWaitForEvent(t, "colreset", &got)
		require.Equal(t, 3, len(got.BoardColumns))
		require.Equal(t, "Appreciations", got.BoardColumns[0].Text)

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Guest user cannot update columns", func(t *testing.T) {
		newCols := []*harness.BoardColumn{
			{Id: "col01", Text: "Hacked", Color: "green", Position: 1, IsDefault: false},
		}
		require.NoError(t, userB.ChangeColumns(newCols))
		require.NoError(t, userA.MustNotReceiveEvent("colreset"))

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Cannot delete columns with existing messages", func(t *testing.T) {
		// Current state has col04, col02, col01
		// Add message to col01.
		msgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)
		require.NoError(t, userA.SendMessage(msgId, "Protect this column", "col01"))
		userA.FlushEvents() // Wait for own message
		userB.FlushEvents()

		// Try to set columns to only col04, col02 (effectively deleting col01)
		newCols := []*harness.BoardColumn{
			{Id: "col04", Text: "Appreciations", Color: "fuchsia", Position: 1, IsDefault: true},
			{Id: "col02", Text: "Stop", Color: "red", Position: 2, IsDefault: false},
		}

		require.NoError(t, userA.ChangeColumns(newCols))
		require.NoError(t, userB.MustNotReceiveEvent("colreset"))

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Column count validation", func(t *testing.T) {
		// Too many columns (> 5)
		newCols := []*harness.BoardColumn{
			{Id: "c1", Text: "1", Color: "c", Position: 1},
			{Id: "c2", Text: "2", Color: "c", Position: 2},
			{Id: "c3", Text: "3", Color: "c", Position: 3},
			{Id: "c4", Text: "4", Color: "c", Position: 4},
			{Id: "c5", Text: "5", Color: "c", Position: 5},
			{Id: "c6", Text: "6", Color: "c", Position: 6},
		}
		require.NoError(t, userA.ChangeColumns(newCols))
		require.NoError(t, userB.MustNotReceiveEvent("colreset"))

		// Empty columns
		require.NoError(t, userA.ChangeColumns([]*harness.BoardColumn{}))
		require.NoError(t, userB.MustNotReceiveEvent("colreset"))

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Column text length validation", func(t *testing.T) {
		// Text > 80 chars
		longText := "This text is definitely longer than eighty characters which is the limit for the column name text so it should fail"
		require.True(t, len([]rune(longText)) > 80)
		newCols := []*harness.BoardColumn{
			{Id: "col01", Text: longText, Color: "green", Position: 1},
		}
		require.NoError(t, userA.ChangeColumns(newCols))
		require.NoError(t, userB.MustNotReceiveEvent("colreset"))

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Cannot update columns in a locked board", func(t *testing.T) {
		// Lock board
		require.NoError(t, userA.LockBoard(true))
		var lockResp harness.LockResponse
		userB.MustWaitForEvent(t, "lock", &lockResp)
		require.True(t, lockResp.Lock)
		userA.FlushEvents()
		userB.FlushEvents()

		newCols := []*harness.BoardColumn{
			{Id: "col02", Text: "Try Update", Color: "red", Position: 1},
		}
		require.NoError(t, userA.ChangeColumns(newCols))
		require.NoError(t, userB.MustNotReceiveEvent("colreset"))

		// Unlock for further tests
		require.NoError(t, userA.LockBoard(false))

		userA.FlushEvents()
		userB.FlushEvents()
	})
}

func TestBoardDeletion(t *testing.T) {
	_, userA, userB := harness.SetupTest(t, true)

	t.Run("Guest user cannot delete board", func(t *testing.T) {
		require.NoError(t, userB.DeleteBoard())
		require.NoError(t, userA.MustNotReceiveEvent("delall"))
	})

	t.Run("Board owner can delete board", func(t *testing.T) {
		require.NoError(t, userA.DeleteBoard())
		var got harness.DeleteAllResponse
		userB.MustWaitForEvent(t, "delall", &got)
		userA.MustWaitForEvent(t, "delall", &got)

		userA.FlushEvents()
		userB.FlushEvents()
	})

	// Negative Test: Registration to non-existent board
	t.Run("Do NOT allow registration to non-existent(deleted) board", func(t *testing.T) {
		// originalBoard := userA.Board
		// userA.Board = "nonexistent-board"

		require.NoError(t, userA.Register())

		require.NoError(t, userA.MustNotReceiveAnyEvent())

		userA.FlushEvents()
		userB.FlushEvents()
		// userA.Board = originalBoard // Restore
	})

	t.Run("New message to non-existent(deleted) board should fail", func(t *testing.T) {
		content := "First message"
		category := "col01"
		msgId := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)

		// originalBoard := userA.Board
		userA.Board = "nonexistent-board"

		require.NoError(t, userA.SendMessage(msgId, content, category))
		require.NoError(t, userA.MustNotReceiveAnyEvent())

		// userA.Board = originalBoard

		userA.FlushEvents()
		userB.FlushEvents()
	})
}

func TestUserLeaving(t *testing.T) {
	_, userA, userB := harness.SetupTest(t, true)

	t.Run("Other users receive closing event when a user leaves", func(t *testing.T) {
		// Bob leaves
		userB.Close()

		var got harness.UserClosingResponse
		userA.MustWaitForEvent(t, "closing", &got)
		require.Equal(t, "xid-"+userB.Id, got.Xid)
	})
}

// Todo: Add tests

// Connection: User attempts to connect to non-existant board should fail
// Message: Updating message in a different board should fail. (Server validates if same msgId that is attached to a board is not "accidently" updated by another user from another board. Add test for that.)
