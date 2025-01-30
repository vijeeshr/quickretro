package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"
)

//go:embed all:frontend/dist/*
var frontendFiles embed.FS

type Config struct {
	Server struct {
		AllowedOrigins []string `toml:"allowed_origins"`
	} `toml:"server"`
	Websocket struct {
		MaxMessageSize int64 `toml:"max_message_size_bytes"`
	} `toml:"websocket"`
	Data struct {
		AutoDeleteDuration string `toml:"auto_delete_duration"`
	} `toml:"data"`
}

var config Config

func main() {
	debug := flag.Bool("debug", false, "set to true to run in debug mode")
	flag.Parse()

	// Prepare Logger
	var logger *slog.Logger
	if *debug {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	} else {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	slog.SetDefault(logger)

	// Load configuration
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		slog.Error("Failed to load configuration from config.toml", "error", err)
		os.Exit(1)
	}

	// Parse Auto-Delete time duration
	autoDeleteDuration, err := parseDuration(config.Data.AutoDeleteDuration)
	if err != nil {
		slog.Error("Invalid auto-delete duration format", "error", err)
		os.Exit(1)
	}

	// Connect to Redis
	ctx := context.Background()
	red := NewRedisConnector(ctx, autoDeleteDuration)
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

func parseDuration(s string) (time.Duration, error) {
	var multiplier time.Duration = 1
	switch {
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
		return 0, fmt.Errorf("invalid duration format: missing unit (use s/m/h/d)")
	}

	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration value: %w", err)
	}

	return time.Duration(value) * multiplier, nil
}
