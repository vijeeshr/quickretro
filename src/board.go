package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

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
	Id           string      `redis:"id"`
	Name         string      `redis:"name"`
	Team         string      `redis:"team"`
	Owner        string      `redis:"owner"`
	Status       BoardStatus `redis:"status"`
	Mask         bool        `redis:"mask"`
	CreatedAtUtc time.Time   `redis:"createdAtUtc"`
}

type BoardColumn struct {
	Id    string `redis:"id" json:"id"`
	Text  string `redis:"text" json:"text"`
	Color string `redis:"color" json:"color"`
}

func (b BoardColumn) String() string {
	return fmt.Sprintf("Id:%s Text:%s Color:%s", b.Id, b.Text, b.Color)
}

type CreateBoardReq struct {
	Name    string         `json:"name"`
	Team    string         `json:"team"`
	Owner   string         `json:"owner"`
	Columns []*BoardColumn `json:"columns"`
}

type CreateBoardRes struct {
	Id string `json:"id"`
}

type GetBoardRes struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	IsOwner bool   `json:"isOwner"`
}

// ToDo: Protect this from abuse. This can be misused for DOS, as it adds data to Redis on every call.
// Creates a new board and returns it
func HandleCreateBoard(c *RedisConnector, w http.ResponseWriter, r *http.Request) {
	// Todo: Protect this function with auth and permissions.
	// Validate request.
	// if r.Method != "POST" {
	// 	w.WriteHeader(http.StatusMethodNotAllowed)
	// 	return
	// }

	// Parse request
	var createReq CreateBoardReq
	err := decodeJSONBody(w, r, &createReq)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			slog.Error("Error parsing CreateBoardRequest", "details", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// Todo: Validate parsed payload
	// name, ok := mux.Vars(r)["name"]
	// if !ok || name == "" {
	// 	// If board is not passed, return as Bad request.
	// 	log.Println("Board name not passed")
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	if len(createReq.Columns) == 0 {
		slog.Error("Columns missing in create board request payload")
		http.Error(w, "Columns missing", http.StatusBadRequest)
		return
	}

	// Start creation
	id := shortuuid.New()
	board := &Board{Id: id, Name: createReq.Name, Team: createReq.Team, Owner: createReq.Owner, Status: InProgress, Mask: true, CreatedAtUtc: time.Now().UTC()}

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

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
		likes := red.GetLikesCountMultiple(ids...)

		for _, m := range msgs {
			msgRes := m.NewResponse("msg").(MessageResponse)
			if count, ok := likes[m.Id]; ok {
				msgRes.Likes = strconv.FormatInt(count, 10)
			}
			msgRes.Mine = m.By == user
			msgRes.Liked = red.HasLiked(m.Id, user) // Todo: This calls Redis SISMEMBER [O(1) as per doc] in a loop. Check for impact.
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
