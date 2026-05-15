package harness

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

const BaseURL string = "http://localhost:8921" // https://localhost

// Constants from source (mirrored here for validaiton)
const (
	MaxIdSizeBytes        int = 36
	MaxColumnIdSizeBytes  int = 5
	MaxColorSizeBytes     int = 24
	MaxCategoryTextLength int = 80
	MaxTextLength         int = 80
)

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

	// resp, err := client.Post(BaseURL+"/api/board/create", "application/json", bytes.NewBuffer(body))
	// require.NoError(t, err)
	req, err := http.NewRequest("POST", BaseURL+"/api/board/create", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://localhost")
	resp, err := client.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	return result["id"]
}

// SetupBoardAndUsers creates a board and three users (Alice as owner, Bob as guest, Clark as guest) but does not connect them.
func SetupBoardAndUsers(t *testing.T) (string, *TestUser, *TestUser, *TestUser) {
	// Shared Setup
	userAId := "user-a"
	userBId := "user-b"
	userCId := "user-c"

	// Create Board
	boardId := CreateBoard(t, userAId)
	t.Logf("Created board: %s", boardId)

	userA := NewUser(userAId, "Alice", boardId)
	userB := NewUser(userBId, "Bob", boardId)
	userC := NewUser(userCId, "Clark", boardId)

	// t.Cleanup(func() {
	// 	userA.Close()
	// 	userB.Close()
	// 	userC.Close()
	// })

	return boardId, userA, userB, userC
}

// SetupTest creates a board and three users (Alice as owner, Bob as guest, Clark as guest),
// connects them.
// If autoRegister is true, it also performs the registration flow for all users.
// It returns the board ID and the two users.
func SetupTest(t *testing.T, autoRegister bool) (string, *TestUser, *TestUser, *TestUser) {
	boardId, userA, userB, userC := SetupBoardAndUsers(t)

	// Connect Users
	require.NoError(t, userA.Connect(BaseURL))
	require.NoError(t, userB.Connect(BaseURL))
	require.NoError(t, userC.Connect(BaseURL))

	t.Cleanup(func() {
		userA.Close()
		userB.Close()
		userC.Close()
	})

	if autoRegister {
		// Register Alice
		require.NoError(t, userA.Register())
		// Register Bob
		require.NoError(t, userB.Register())
		// Register Clark
		require.NoError(t, userC.Register())

		// Flush any extra events to ensure clean slate
		userA.FlushEvents()
		userB.FlushEvents()
		userC.FlushEvents()
	}

	return boardId, userA, userB, userC
}
