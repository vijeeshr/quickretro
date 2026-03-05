package scenarios

import (
	"e2e_tests/harness"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConnectHandshake(t *testing.T) {

	_, userA, userB := harness.SetupBoardAndUsers(t)

	t.Run("Xid should remain same after calling connect multiple times", func(t *testing.T) {
		var firstReg harness.RegisterResponse
		var secondReg harness.RegisterResponse
		var joinRes harness.UserJoiningResponse

		// Initial connect handshake for both users
		require.NoError(t, userA.Connect(harness.BaseURL))
		require.NoError(t, userB.Connect(harness.BaseURL))

		// Check xid value in reg response
		require.NoError(t, userA.Register())
		userA.MustWaitForEvent(t, "reg", &firstReg)
		require.NotEmpty(t, firstReg.Xid, "Xid should not be empty")
		//
		userB.MustWaitForEvent(t, "joining", &joinRes)
		require.Equal(t, firstReg.Xid, joinRes.Xid, "Xid in joining and register response should be same")

		// Re-connect UserA
		// userA.Close() (To use close(), uncomment // u.Done = make(chan struct{}) in user.go)
		require.NoError(t, userA.Connect(harness.BaseURL))

		// Register UserA again
		require.NoError(t, userA.Register())
		userA.MustWaitForEvent(t, "reg", &secondReg)

		// Expect Xid to be exactly the same
		require.NotEmpty(t, secondReg.Xid, "Second Xid should not be empty")
		require.Equal(t, secondReg.Xid, firstReg.Xid, "Xid should remain the same across reconnects for same board")

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("Users should be alloted different xids", func(t *testing.T) {
		var regRespUserA harness.RegisterResponse
		var regRespUserB harness.RegisterResponse

		// Check xid value in reg response
		require.NoError(t, userA.Register())
		userA.MustWaitForEvent(t, "reg", &regRespUserA)
		require.NoError(t, userB.Register())
		userB.MustWaitForEvent(t, "reg", &regRespUserB)

		require.NotEmpty(t, regRespUserA.Xid, "UserA.Xid should not be empty")
		require.NotEmpty(t, regRespUserB.Xid, "UserB.Xid should not be empty")
		require.NotEqual(t, regRespUserA.Xid, regRespUserB, "Xid should NOT be same for different users")

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("UserA Xid nickname change should reflect in new messages", func(t *testing.T) {

		var regRes harness.RegisterResponse
		var firstMsg harness.MessageResponse
		var secondMsg harness.MessageResponse

		originalNickname := "Alice"
		newNickname := "Alice Edwards"
		require.Equal(t, originalNickname, userA.Nickname)

		// Ensure userB has joined and registered to receive responses
		require.NoError(t, userB.Register())
		userA.FlushEvents()
		userB.FlushEvents()

		// Register and create message
		require.NoError(t, userA.Register())
		userA.MustWaitForEvent(t, "reg", &regRes)
		var joiningRes harness.UserJoiningResponse
		userB.MustWaitForEvent(t, "joining", &joiningRes)
		msgId1 := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)
		require.NoError(t, userA.SendMessage(msgId1, "Hello 1", "col01"))
		userA.MustWaitForEvent(t, "msg", &firstMsg)

		// UserB can see nickname of UserA in "joining" response
		require.Equal(t, userA.Nickname, joiningRes.Nickname)

		// UserA changes nickname on reconnect
		// userA.Close() (To use close(), uncomment // u.Done = make(chan struct{}) in user.go)
		userA.Nickname = newNickname
		require.NoError(t, userA.Connect(harness.BaseURL))

		// Create new message
		msgId2 := fmt.Sprintf("msg-%d-%s", time.Now().UnixNano(), userA.Id)
		require.NoError(t, userA.SendMessage(msgId2, "Hello 2", "col01"))
		userA.MustWaitForEvent(t, "msg", &secondMsg)

		require.Equal(t, originalNickname, firstMsg.ByNickname, "First message should have original nickname")
		require.Equal(t, newNickname, secondMsg.ByNickname, "Second message should have updated nickname")

		userA.FlushEvents()
		userB.FlushEvents()

		// Register again
		require.NoError(t, userA.Register())
		userA.MustWaitForEvent(t, "reg", &regRes)
		userB.MustWaitForEvent(t, "joining", &joiningRes)
		// UserA should see new nickname in "reg" response
		require.ElementsMatch(t, []harness.UserDetails{
			{Nickname: newNickname, Xid: "1"},
			{Nickname: userB.Nickname, Xid: "2"},
		}, regRes.Users, "Reg response should show new nickname")
		// UserB should see new nickname of UserA in "joining" response
		require.Equal(t, newNickname, joiningRes.Nickname, "Joining response should show new nickname")

		userA.FlushEvents()
		userB.FlushEvents()
	})
}
