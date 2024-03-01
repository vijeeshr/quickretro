package main

import (
	"context"
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

//go:embed all:frontend/dist/*
var frontendFiles embed.FS

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Connect to Redis
	ctx := context.Background()
	red := NewRedisConnector(ctx)
	defer red.Close()

	// Prepare Hub
	hub := newHub(red)
	go hub.run()

	// Setup routes and handlers
	// Todo: Check origin
	router := mux.NewRouter()

	router.HandleFunc("/api/board/create", func(w http.ResponseWriter, r *http.Request) {
		HandleCreateBoard(red, w, r)
	}).Methods("POST") // Todo: Check origin in handler

	router.HandleFunc("/api/board/{id}/user/{user}", func(w http.ResponseWriter, r *http.Request) {
		HandleGetBoard(red, w, r)
	}).Methods("GET") // Todo: Check origin in handler

	router.HandleFunc("/api/board/{id}/user/{user}/refresh", func(w http.ResponseWriter, r *http.Request) {
		HandleRefresh(red, w, r)
	}).Methods("GET") // Todo: Check origin in handler

	router.HandleFunc("/ws/board/{board}/user/{user}/meet", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(hub, w, r)
	})

	// Serve static files from the embedded file system
	assetsFS, _ := fs.Sub(frontendFiles, "frontend/dist/assets")
	assetsHandler := http.FileServer(http.FS(assetsFS))
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", assetsHandler)).Methods("GET")
	router.HandleFunc("/create", frontendIndexHandler).Methods("GET")
	router.HandleFunc("/board/{id}/join", frontendIndexHandler).Methods("GET")
	router.HandleFunc("/board/{id}/", frontendIndexHandler).Methods("GET")
	router.HandleFunc("/board/{id}", frontendIndexHandler).Methods("GET")
	router.HandleFunc("/", frontendIndexHandler).Methods("GET")

	//err := http.ListenAndServe(":8080", nil)
	logger.Info("Server listening on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func frontendIndexHandler(w http.ResponseWriter, r *http.Request) {
	// http.ServeFileFS ?
	indexFile, err := frontendFiles.ReadFile("frontend/dist/index.html")
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Write(indexFile)
}
