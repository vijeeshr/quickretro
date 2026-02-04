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

	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"
)

//go:embed all:frontend/dist/*
var frontendFiles embed.FS

// Constants for input validation
const (
	MaxIdSizeBytes       int = 36 // (UUIDs, shortUUIDs). These are ASCII-only, machine-generated values used to validate inputs for BoardId, UserId, Xid
	MaxColumnIdSizeBytes int = 5
	MaxColorSizeBytes    int = 24
)

type Config struct {
	Server struct {
		TurnstileSiteVerifyUrl string   `toml:"turnstile_site_verify_url"`
		AllowedOrigins         []string `toml:"allowed_origins"`
	} `toml:"server"`
	Data struct {
		AutoDeleteDuration    string `toml:"auto_delete_duration"`
		MaxCategoryTextLength int    `toml:"max_category_text_length"`
		MaxTextLength         int    `toml:"max_text_length"`
	} `toml:"data"`
	Websocket struct {
		MaxMessageSizeBytes int64 `toml:"max_message_size_bytes"`
	} `toml:"websocket"`
	Frontend struct {
		ContentEditableInvalidDebounceMs uint16 `toml:"content_editable_invalid_debounce_ms"`
	} `toml:"frontend"`
}

type EnvironmentConfig struct {
	RedisConnStr       string
	TurnstileSiteKey   string
	TurnstileSecretKey string
	TurnstileEnabled   bool
}

var config Config
var envConfig EnvironmentConfig

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

	// Load Environment configuration
	envConfig = LoadEnvironmentConfig()
	// slog.Info("Loaded environment configuration", "TurnstileEnabled", envConfig.TurnstileEnabled)

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

	// router.HandleFunc("/api/board/{id}/user/{user}", func(w http.ResponseWriter, r *http.Request) {
	// 	HandleGetBoard(red, w, r)
	// }).Methods("GET") // Todo: Check origin in handler

	// router.HandleFunc("/api/board/{id}/user/{user}/refresh", func(w http.ResponseWriter, r *http.Request) {
	// 	HandleRefresh(red, w, r)
	// }).Methods("GET") // Todo: Check origin in handler

	router.HandleFunc("/ws/board/{board}/user/{user}/meet", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(hub, w, r)
	})

	// Serve static files from the embedded file system
	assetsFS, _ := fs.Sub(frontendFiles, "frontend/dist/assets")
	assetsHandler := http.FileServer(http.FS(assetsFS))
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", assetsHandler)).Methods("GET")
	router.HandleFunc("/config.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		turnstileEnabled := envConfig.TurnstileEnabled
		turnstileSiteKey := envConfig.TurnstileSiteKey

		js := fmt.Sprintf(`window.APP_CONFIG = {
		turnstile:{enabled:%t,siteKey:"%s"},
		data:{maxCategoryTextLength:%d,maxTextLength:%d},
		websocket:{maxMessageSizeBytes:%d},
		frontend:{contentEditableInvalidDebounceMs:%d}
		};`,
			turnstileEnabled,
			turnstileSiteKey,
			config.Data.MaxCategoryTextLength,
			config.Data.MaxTextLength,
			config.Websocket.MaxMessageSizeBytes,
			config.Frontend.ContentEditableInvalidDebounceMs,
		)

		_, _ = w.Write([]byte(js))
	}).Methods("GET")

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

func LoadEnvironmentConfig() EnvironmentConfig {

	// The values to "fallback" arg of getEnv is treated as default when running the app outside of Docker (with Compose)
	// This is usually the case during local development, as the ENV vars may not exist.
	// So for local development, you can modify the values as per your need.

	// Note for Redis
	// --------------
	// The default "redis://localhost:6379/0" is for accessing redis from host, during development, when running the app locally i.e. not within Docker.
	// This default may fail when running inside a Docker container as localhost inside a container refers to itself.
	// So, ensure REDIS_CONNSTR environment variable is correctly set.
	// redisConnStr := getEnv("REDIS_CONNSTR", "redis://app-user:mysecretpassword@localhost:6379/0") // Pattern for ACL from local

	// Cloudflare Turnstile Dummy SiteKeys and SecretKeys for development
	// Sitekey					Description	                    Visibility
	// -------                  -----------                     ----------
	// 1x00000000000000000000AA	Always passes	                visible
	// 2x00000000000000000000AB	Always blocks	                visible
	// 1x00000000000000000000BB	Always passes	                invisible
	// 2x00000000000000000000BB	Always blocks	                invisible
	// 3x00000000000000000000FF	Forces an interactive challenge	visible

	// SecretKey							Result
	// ---------							------
	// 1x0000000000000000000000000000000AA	Always passes
	// 2x0000000000000000000000000000000AA	Always fails
	// 3x0000000000000000000000000000000AA	Yields a "token already spent" error
	// https://developers.cloudflare.com/turnstile/troubleshooting/testing/

	return EnvironmentConfig{
		RedisConnStr:       getEnv("REDIS_CONNSTR", "redis://localhost:6379/0"),
		TurnstileEnabled:   getEnv("TURNSTILE_ENABLED", "false") == "true",
		TurnstileSiteKey:   getEnv("TURNSTILE_SITE_KEY", "1x00000000000000000000AA"),
		TurnstileSecretKey: getEnv("TURNSTILE_SECRET_KEY", "1x0000000000000000000000000000000AA"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
