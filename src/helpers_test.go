package main

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_decodeJSONBody(t *testing.T) {
	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name           string
		contentType    string
		body           string
		errMsg         string
		expectedStatus int
		expectErr      bool
	}{
		{
			name:           "Valid JSON",
			contentType:    "application/json",
			body:           `{"name": "Alice", "age": 30}`,
			expectedStatus: 0, // No error
			expectErr:      false,
		},
		{
			name:           "Wrong Content-Type",
			contentType:    "text/plain",
			body:           `{"name": "Alice"}`,
			expectedStatus: http.StatusUnsupportedMediaType,
			expectErr:      true,
			errMsg:         "Content-Type header is not application/json",
		},
		{
			name:           "Malformed JSON (Syntax Error)",
			contentType:    "application/json",
			body:           `{"name": "Alice"`, // Missing closing brace
			expectedStatus: http.StatusBadRequest,
			expectErr:      true,
			errMsg:         "Request body contains badly-formed JSON",
		},
		{
			name:           "Invalid Value Type",
			contentType:    "application/json",
			body:           `{"name": "Alice", "age": "thirty"}`, // age should be int
			expectedStatus: http.StatusBadRequest,
			expectErr:      true,
			errMsg:         "Request body contains an invalid value for the \"age\" field",
		},
		{
			name:           "Unknown Field",
			contentType:    "application/json",
			body:           `{"name": "Alice", "city": "London"}`,
			expectedStatus: http.StatusBadRequest,
			expectErr:      true,
			errMsg:         "Request body contains unknown field",
		},
		{
			name:           "Empty Body",
			contentType:    "application/json",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectErr:      true,
			errMsg:         "Request body must not be empty",
		},
		{
			name:           "Multiple JSON Objects",
			contentType:    "application/json",
			body:           `{"name": "Alice"}{"name": "Bob"}`,
			expectedStatus: http.StatusBadRequest,
			expectErr:      true,
			errMsg:         "Request body must only contain a single JSON object",
		},
		{
			name:           "Body Too Large",
			contentType:    "application/json",
			body:           "[" + strings.Repeat("0,", 1048576) + "0]", // Valid JSON array: [0,0,0,0...] which exceeds 1MB
			expectedStatus: http.StatusRequestEntityTooLarge,
			expectErr:      true,
			errMsg:         "Request body must not be larger than 1MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup request and recorder
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rr := httptest.NewRecorder()

			var dst testStruct
			err := decodeJSONBody(rr, req, &dst)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected an error but got nil")
				}

				// Check if it's our custom malformedRequest type
				if mr, ok := errors.AsType[*malformedRequest](err); ok {
					if mr.status != tt.expectedStatus {
						t.Errorf("expected status %d, got %d", tt.expectedStatus, mr.status)
					}
					if !strings.Contains(mr.msg, tt.errMsg) {
						t.Errorf("expected error message to contain %q, got %q", tt.errMsg, mr.msg)
					}
				} else {
					t.Errorf("expected error to be of type *malformedRequest, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func Test_parseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"10s", 10 * time.Second, false},
		{"5m", 5 * time.Minute, false},
		{"2h", 2 * time.Hour, false},
		{"3d", 3 * 24 * time.Hour, false},
		{"0s", 0, false},
		{"100m", 100 * time.Minute, false},
		{"-5m", -5 * time.Minute, false},
		{"abc", 0, true},
		{"10x", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseDuration(tt.input)
			if (err != nil) != tt.hasError {
				t.Errorf("parseDuration(%q) error = %v, wantErr %v", tt.input, err, tt.hasError)
			}
			if result != tt.expected {
				t.Errorf("parseDuration(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
