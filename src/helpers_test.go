package main

import (
	"testing"
	"time"
)

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
