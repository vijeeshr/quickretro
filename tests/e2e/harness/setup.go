package harness

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"

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

// SetupTest creates a board and two users (Alice as owner, Bob as guest),
// connects them.
// If autoRegister is true, it also performs the registration flow for both users.
// It returns the board ID and the two users.
func SetupTest(t *testing.T, autoRegister bool) (string, *TestUser, *TestUser) {
	// Shared Setup
	userAId := "user-a"
	userBId := "user-b"

	// Create Board
	boardId := CreateBoard(t, userAId)
	t.Logf("Created board: %s", boardId)

	userA := NewUser(userAId, "Alice", boardId)
	userB := NewUser(userBId, "Bob", boardId)

	// Connect Users
	require.NoError(t, userA.Connect(BaseURL))
	require.NoError(t, userB.Connect(BaseURL))

	t.Cleanup(func() {
		userA.Close()
		userB.Close()
	})

	if autoRegister {
		// Register Alice
		require.NoError(t, userA.Register())
		// // Alice receives own reg
		// userA.MustWaitForEvent(t, "reg", nil)
		// // Bob receives joining
		// userB.MustWaitForEvent(t, "joining", nil)

		// Register Bob
		require.NoError(t, userB.Register())
		// userB.MustWaitForEvent(t, "reg", nil)
		// // Alice receives joining for Bob
		// userA.MustWaitForEvent(t, "joining", nil)

		// Flush any extra events to ensure clean slate
		userA.FlushEvents()
		userB.FlushEvents()
	}

	return boardId, userA, userB
}
