package main

// https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body

import (
	"encoding/json/jsontext"
	json "encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type malformedRequest struct {
	msg    string
	status int
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) error {
	// Check Content-Type
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: "Content-Type header is not application/json"}
		}
	}

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := jsontext.NewDecoder(r.Body)

	err := json.UnmarshalDecode(dec, dst, json.RejectUnknownMembers(true))
	if err != nil {
		var msg string
		var status int = http.StatusBadRequest

		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			msg = "Request body must not be larger than 1MB"
			status = http.StatusRequestEntityTooLarge
		} else if syntacticError, ok := errors.AsType[*jsontext.SyntacticError](err); ok {
			msg = fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntacticError.ByteOffset)
		} else if errors.Is(err, io.ErrUnexpectedEOF) {
			msg = "Request body contains badly-formed JSON"
		} else if semanticError, ok := errors.AsType[*json.SemanticError](err); ok {
			if errors.Is(semanticError.Err, json.ErrUnknownName) {
				msg = fmt.Sprintf("Request body contains unknown field %s", semanticError.JSONPointer.LastToken())
			} else {
				fieldName := semanticError.JSONPointer.LastToken()
				if fieldName != "" {
					msg = fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", fieldName, semanticError.ByteOffset)
				} else {
					msg = fmt.Sprintf("Request body contains an invalid value (at position %d)", semanticError.ByteOffset)
				}
			}
		} else if errors.Is(err, io.EOF) {
			msg = "Request body must not be empty"
		} else {
			return err
		}

		return &malformedRequest{status: status, msg: msg}
	}

	// Ensure there is no trailing data
	if _, err = dec.ReadToken(); !errors.Is(err, io.EOF) {
		return &malformedRequest{status: http.StatusBadRequest, msg: "Request body must only contain a single JSON object"}
	}

	return nil
}

func parseDuration(s string) (time.Duration, error) {
	var multiplier time.Duration = 1
	switch {
	case strings.HasSuffix(s, "ms"):
		multiplier = time.Millisecond
		s = strings.TrimSuffix(s, "ms")
	case strings.HasSuffix(s, "s"):
		multiplier = time.Second
		s = strings.TrimSuffix(s, "s")
	case strings.HasSuffix(s, "m"):
		multiplier = time.Minute
		s = strings.TrimSuffix(s, "m")
	case strings.HasSuffix(s, "h"):
		multiplier = time.Hour
		s = strings.TrimSuffix(s, "h")
	case strings.HasSuffix(s, "d"):
		multiplier = 24 * time.Hour
		s = strings.TrimSuffix(s, "d")
	default:
		return 0, fmt.Errorf("invalid duration format: missing unit (use ms/s/m/h/d)")
	}

	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration value: %w", err)
	}

	return time.Duration(value) * multiplier, nil
}
