package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
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

	router.HandleFunc("/board/{id}/meet", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(hub, w, r)
	})

	// Static files

	// router.PathPrefix("/board/{id}/").Handler(http.StripPrefix("/board/", http.FileServer(http.Dir("public"))))
	// router.HandleFunc("/board/{id}", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "public/index.html")
	// }).Methods("GET")

	// router.PathPrefix("/board/").Handler(http.StripPrefix("/board/", http.FileServer(http.Dir("public"))))
	// router.HandleFunc("/board", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "public/index.html")
	// })

	router.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/create.html")
	})

	router.HandleFunc("/board/{id}/join", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/join.html")
	})

	router.HandleFunc("/board/script.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/script.js")
	})
	router.HandleFunc("/board/output.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/output.css")
	})
	router.HandleFunc("/board/{id}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/board.html")
	})

	fh := http.FileServer(http.Dir("public"))
	router.PathPrefix("/board/").Handler(http.StripPrefix("/board/", fh))
	router.PathPrefix("/board").Handler(http.StripPrefix("/board", fh))
	router.PathPrefix("/").Handler(http.StripPrefix("/", fh))

	//err := http.ListenAndServe(":8080", nil)
	log.Println("Server listening on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil { // Todo: use https/wss
		log.Fatal(err)
	}
}
