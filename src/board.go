package main

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"unicode/utf8"

	"github.com/gorilla/mux"
	"github.com/lithammer/shortuuid/v4"
)

type BoardStatus int

const (
	InProgress BoardStatus = iota
	Paused
	Completed
)

func (b BoardStatus) String() string {
	return [...]string{"inProgress", "paused", "completed"}[b]
}

type Board struct {
	Id                string      `redis:"id"`
	Name              string      `redis:"name"`
	Team              string      `redis:"team"`
	Owner             string      `redis:"owner"`
	Creator           string      `redis:"creator"`
	Status            BoardStatus `redis:"status"`
	Mask              bool        `redis:"mask"`
	Lock              bool        `redis:"lock"`
	TimerExpiresAtUtc int64       `redis:"timerExpiresAtUtc"`
	CreatedAtUtc      int64       `redis:"createdAtUtc"`
	AutoDeleteAtUtc   int64       `redis:"autoDeleteAtUtc"`
}

type BoardColumn struct {
	Id        string `redis:"id" json:"id"`
	Text      string `redis:"text" json:"text"`
	Color     string `redis:"color" json:"color"`
	Position  int    `redis:"pos" json:"pos"`
	IsDefault bool   `redis:"isDefault" json:"isDefault"`
}

func (b BoardColumn) String() string {
	return fmt.Sprintf("Id:%s Text:%s Color:%s Position:%d", b.Id, b.Text, b.Color, b.Position)
}

type CreateBoardReq struct {
	Name                string         `json:"name"`
	Team                string         `json:"team"`
	Owner               string         `json:"owner"`
	CfTurnstileResponse string         `json:"cfTurnstileResponse"`
	Columns             []*BoardColumn `json:"columns"`
}

type CreateBoardRes struct {
	Id string `json:"id"`
}

type GetBoardRes struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	IsOwner bool   `json:"isOwner"`
}

// Check if the request Origin header is in the configured allowed origins list.
func isOriginAllowed(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	return origin != "" && slices.Contains(config.Server.AllowedOrigins, origin)
}

// Creates a new board and returns it
func HandleCreateBoard(c *RedisConnector, w http.ResponseWriter, r *http.Request) {
	// Validate Origin
	if !isOriginAllowed(r) {
		slog.Warn("Rejected request with disallowed origin", "origin", r.Header.Get("Origin"), "remote", r.RemoteAddr)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	// Parse request
	var createReq CreateBoardReq
	err := decodeJSONBody(w, r, &createReq)
	if err != nil {
		if mr, ok := errors.AsType[*malformedRequest](err); ok {
			http.Error(w, mr.msg, mr.status)
		} else {
			slog.Error("Error parsing CreateBoardRequest", "details", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// Add Turnstile validation
	if envConfig.TurnstileEnabled {
		if createReq.CfTurnstileResponse == "" {
			http.Error(w, "CAPTCHA verification required", http.StatusBadRequest)
			return
		}

		// Get client IP (consider X-Forwarded-For if behind proxy)
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			slog.Error("Error parsing remote address", "error", err)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		valid, err := verifyTurnstile(createReq.CfTurnstileResponse, ip)
		if err != nil || !valid {
			slog.Warn("Turnstile verification failed", "error", err, "ip", ip)
			http.Error(w, "CAPTCHA verification failed", http.StatusBadRequest)
			return
		}
	}

	// Todo: Validate parsed payload
	// name, ok := mux.Vars(r)["name"]
	// if !ok || name == "" {
	// 	// If board is not passed, return as Bad request.
	// 	log.Println("Board name not passed")
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	if len(createReq.Columns) == 0 || len(createReq.Columns) > 5 {
		slog.Error("Invalid Columns data in create board request payload")
		http.Error(w, "Invalid columns data", http.StatusBadRequest)
		return
	}

	for _, col := range createReq.Columns {
		if col == nil {
			slog.Error("CreateBoardRequest contains nil column definition")
			http.Error(w, "Invalid columns data", http.StatusBadRequest)
			return
		}
		textLen := utf8.RuneCountInString(col.Text)
		if len(col.Id) > MaxColumnIdSizeBytes || len(col.Color) > MaxColorSizeBytes || textLen > config.Data.MaxCategoryTextLength {
			slog.Error("Column info exceeds limit in create board request payload", "col", col.Id, "len", textLen, "len-color", len(col.Color))
			http.Error(w, "Column info exceeds limit", http.StatusBadRequest)
			return
		}
	}

	if utf8.RuneCountInString(createReq.Name) > config.Data.MaxTextLength {
		slog.Error("Board name exceeds limit in create board request payload")
		http.Error(w, "Board name exceeds length limit", http.StatusBadRequest)
		return
	}

	if utf8.RuneCountInString(createReq.Team) > config.Data.MaxTextLength {
		slog.Error("Team name exceeds limit in create board request payload")
		http.Error(w, "Team name exceeds length limit", http.StatusBadRequest)
		return
	}

	// Owner is userId(UUIDv4) with 36 bytes
	if len(createReq.Owner) > MaxIdSizeBytes {
		slog.Error("OwnerId exceeds limit in create board request payload")
		http.Error(w, "OwnerId exceeds length limit", http.StatusBadRequest)
		return
	}

	// Start creation
	id := shortuuid.New()
	board := &Board{Id: id, Name: createReq.Name, Team: createReq.Team, Owner: createReq.Owner, Creator: createReq.Owner, Status: InProgress, Lock: false, Mask: true}

	// Save to Redis
	if ok := c.CreateBoard(board, createReq.Columns); !ok {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(CreateBoardRes{Id: board.Id})
	if err != nil {
		slog.Error("Error marshalling CreateBoardRes", "details", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	slog.Info("Created", "board", board.Id, "owner", board.Owner)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

func verifyTurnstile(token, remoteIP string) (bool, error) {
	if envConfig.TurnstileSecretKey == "" {
		return false, fmt.Errorf("TURNSTILE_SECRET_KEY not configured")
	}

	data := url.Values{}
	data.Set("secret", envConfig.TurnstileSecretKey)
	data.Set("response", token)
	data.Set("remoteip", remoteIP)

	resp, err := http.PostForm(config.Server.TurnstileSiteVerifyUrl, data)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Success, nil
}

// Deprecated: No longer used
// Returns board by id
func HandleGetBoard(c *RedisConnector, w http.ResponseWriter, r *http.Request) {

	// Validate request
	id, ok := mux.Vars(r)["id"]
	if !ok || id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, ok := mux.Vars(r)["user"]
	if !ok || user == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	board, ok := c.GetBoard(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Prepare response
	data, err := json.Marshal(GetBoardRes{Id: board.Id, Name: board.Name, IsOwner: user == board.Owner})
	if err != nil {
		slog.Error("Error marshalling GetBoardRes", "details", err.Error(), "board", board)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// Deprecated: No longer used
// Returns messages by board
func HandleRefresh(red *RedisConnector, w http.ResponseWriter, r *http.Request) {
	// Todo: Validate properly
	// Validate request.
	id, ok := mux.Vars(r)["id"]
	if !ok || id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, ok := mux.Vars(r)["user"] // Todo: Validate by user
	if !ok || user == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// c := make(chan []MessageResponse)
	// req := &Refresh{group: id, user: user, returnCh: c}

	// hub.refresh <- req
	// res := <-c
	// close(c)

	res := make([]MessageResponse, 0) // End json result is [] instead of null when no messages are present. If var res []MessageResponse was used, [] won't be returned.
	if msgs, ok := red.GetMessages(id); ok {
		// Collect "like" count for all messages in one call via a Redis pipeline
		ids := make([]string, 0)
		for _, m := range msgs {
			ids = append(ids, m.Id)
		}
		likesInfo, likesOk := red.GetLikesInfo(user, ids...)

		for _, m := range msgs {
			msgRes := m.NewMessageResponse()
			msgRes.Mine = m.By == user
			if likesOk {
				if info, ok := likesInfo[m.Id]; ok {
					msgRes.Likes = info.Count
					msgRes.Liked = info.Liked
				}
			}
			res = append(res, msgRes)
		}
	}

	// Prepare response
	data, err := json.Marshal(res)
	if err != nil {
		slog.Error("Error marshalling response for messages", "details", err.Error(), "response", res)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

type AdminBoardInfo struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Team            string `json:"team"`
	Owner           string `json:"owner"`
	Creator         string `json:"creator"`
	CreatedAtUtc    int64  `json:"createdAtUtc"`
	AutoDeleteAtUtc int64  `json:"autoDeleteAtUtc"`
	TTLSeconds      int64  `json:"ttlSeconds"`
}

type AdminBoardsRes struct {
	Boards     []AdminBoardInfo `json:"boards"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"totalPages"`
}

type AdminActionReq struct {
	BoardId string `json:"boardId"`
	Passkey string `json:"passkey"`
}

func validateAdminPasskey(r *http.Request, bodyPasskey string) bool {
	configuredPasskey := envConfig.AdminPasskey
	if configuredPasskey == "" {
		return false
	}
	providedPasskey := r.Header.Get("X-Admin-Passkey")
	if providedPasskey == "" {
		providedPasskey = bodyPasskey
	}
	if providedPasskey == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(providedPasskey), []byte(configuredPasskey)) == 1
}

func HandleAdminVerify(w http.ResponseWriter, r *http.Request) {
	var bodyReq AdminActionReq
	_ = json.NewDecoder(r.Body).Decode(&bodyReq)

	if !validateAdminPasskey(r, bodyReq.Passkey) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func HandleAdminGetBoards(c *RedisConnector, w http.ResponseWriter, r *http.Request) {
	if !validateAdminPasskey(r, "") {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	searchQuery := r.URL.Query().Get("q")
	pageStr := r.URL.Query().Get("page")
	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	limit := 10
	boards, total, err := c.GetBoardsPaginated(searchQuery, page, limit)
	if err != nil {
		slog.Error("Error fetching paginated boards for admin", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	res := AdminBoardsRes{
		Boards:     boards,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func HandleAdminExtendExpiry(c *RedisConnector, w http.ResponseWriter, r *http.Request) {
	var req AdminActionReq
	_ = decodeJSONBody(w, r, &req)

	if !validateAdminPasskey(r, req.Passkey) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if req.BoardId == "" {
		http.Error(w, "boardId required", http.StatusBadRequest)
		return
	}

	info, err := c.ExtendBoardExpiry(req.BoardId, c.timeToLive)
	if err != nil {
		slog.Error("Failed to extend board expiry", "boardId", req.BoardId, "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(info)
}

func HandleAdminRemoveExpiry(c *RedisConnector, w http.ResponseWriter, r *http.Request) {
	var req AdminActionReq
	_ = decodeJSONBody(w, r, &req)

	if !validateAdminPasskey(r, req.Passkey) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if req.BoardId == "" {
		http.Error(w, "boardId required", http.StatusBadRequest)
		return
	}

	info, err := c.RemoveBoardExpiry(req.BoardId)
	if err != nil {
		slog.Error("Failed to remove board expiry", "boardId", req.BoardId, "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(info)
}

func HandleAdminDeleteBoard(c *RedisConnector, w http.ResponseWriter, r *http.Request) {
	var req AdminActionReq
	_ = decodeJSONBody(w, r, &req)

	if !validateAdminPasskey(r, req.Passkey) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if req.BoardId == "" {
		http.Error(w, "boardId required", http.StatusBadRequest)
		return
	}

	if ok := c.DeleteAll(req.BoardId); !ok {
		slog.Error("Failed to delete board via admin", "boardId", req.BoardId)
		http.Error(w, "Failed to delete board", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "boardId": req.BoardId})
}
