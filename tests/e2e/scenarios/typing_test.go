package scenarios

import (
	"e2e_tests/harness"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTypingIndicator(t *testing.T) {
	_, userA, userB := harness.SetupTest(t, true)

	t.Run("User A typing should be broadcasted to User B", func(t *testing.T) {
		// User A sends typing event
		require.NoError(t, userA.SendTyping())

		// User B receives typing event
		var got harness.TypedResponse
		userB.MustWaitForEvent(t, "t", &got)
		require.Equal(t, "xid-"+userA.Id, got.Xid)

		// User A should NOT receive their own typing event
		require.NoError(t, userA.MustNotReceiveEvent("t"))

		userA.FlushEvents()
		userB.FlushEvents()
	})

	t.Run("User B typing should be broadcasted to User A", func(t *testing.T) {
		// User B sends typing event
		require.NoError(t, userB.SendTyping())

		// User A receives typing event
		var got harness.TypedResponse
		userA.MustWaitForEvent(t, "t", &got)
		require.Equal(t, "xid-"+userB.Id, got.Xid)

		// User B should NOT receive their own typing event
		require.NoError(t, userB.MustNotReceiveEvent("t"))

		userA.FlushEvents()
		userB.FlushEvents()
	})
}
