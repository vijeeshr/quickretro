package scenarios

import (
	"bytes"
	"crypto/tls"
	"e2e_tests/harness"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBoardCreation(t *testing.T) {
	// Helper to send create request
	sendCreateRequest := func(t *testing.T, payload map[string]any) (*http.Response, map[string]any) {
		body, _ := json.Marshal(payload)
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}

		resp, err := client.Post(harness.BaseURL+"/api/board/create", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)

		// Parse response if JSON
		var result map[string]any
		if resp.Header.Get("Content-Type") == "application/json" {
			_ = json.NewDecoder(resp.Body).Decode(&result)
		}
		return resp, result
	}

	validColumns := []map[string]any{
		{"id": "col01", "text": "What went well", "isDefault": true, "color": "green", "pos": 1},
		{"id": "col02", "text": "Challenges", "isDefault": true, "color": "red", "pos": 2},
		{"id": "col03", "text": "Action Items", "isDefault": true, "color": "yellow", "pos": 3},
	}

	validPayload := func() map[string]any {
		return map[string]any{
			"name":                "Valid Board",
			"team":                "Valid Team",
			"owner":               "user-valid-owner",
			"cfTurnstileResponse": "", // Disabled in dev/test
			"columns":             validColumns,
		}
	}

	t.Run("Valid board creation", func(t *testing.T) {
		payload := validPayload()
		resp, result := sendCreateRequest(t, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)
		require.NotEmpty(t, result["id"])
	})

	t.Run("Validation: Missing columns", func(t *testing.T) {
		payload := validPayload()
		payload["columns"] = []map[string]any{} // Empty

		resp, _ := sendCreateRequest(t, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Validation: Too many columns", func(t *testing.T) {
		payload := validPayload()
		cols := make([]map[string]any, 6)
		for i := 0; i < 6; i++ {
			cols[i] = validColumns[0] // duplicate is fine for this check essentially
		}
		payload["columns"] = cols

		resp, _ := sendCreateRequest(t, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Validation: Column ID too long", func(t *testing.T) {
		payload := validPayload()
		cols := []map[string]any{
			{"id": strings.Repeat("a", harness.MaxColumnIdSizeBytes+1), "text": "Text", "color": "red", "pos": 1},
		}
		payload["columns"] = cols

		resp, _ := sendCreateRequest(t, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Validation: Column Color too long", func(t *testing.T) {
		payload := validPayload()
		cols := []map[string]any{
			{"id": "c1", "text": "Text", "color": strings.Repeat("a", harness.MaxColorSizeBytes+1), "pos": 1},
		}
		payload["columns"] = cols

		resp, _ := sendCreateRequest(t, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Validation: Column Text too long", func(t *testing.T) {
		payload := validPayload()
		cols := []map[string]any{
			{"id": "c1", "text": strings.Repeat("a", harness.MaxCategoryTextLength+1), "color": "red", "pos": 1},
		}
		payload["columns"] = cols

		resp, _ := sendCreateRequest(t, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Validation: Board Name too long", func(t *testing.T) {
		payload := validPayload()
		payload["name"] = strings.Repeat("a", harness.MaxTextLength+1)

		resp, _ := sendCreateRequest(t, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Validation: Team Name too long", func(t *testing.T) {
		payload := validPayload()
		payload["team"] = strings.Repeat("a", harness.MaxTextLength+1)

		resp, _ := sendCreateRequest(t, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Validation: Owner ID too long", func(t *testing.T) {
		payload := validPayload()
		payload["owner"] = strings.Repeat("a", harness.MaxIdSizeBytes+1)

		resp, _ := sendCreateRequest(t, payload)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
