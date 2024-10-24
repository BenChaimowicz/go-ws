package main

import (
	"errors"
	"io"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	setupServer(mux)

	err := http.ListenAndServe(":5000", mux)

	if errors.Is(err, http.ErrServerClosed) {
		log.Println("Server closed.")
	} else {
		log.Println(err)
	}
}

func setupServer(m *http.ServeMux) {
	hub := NewHub()

	m.HandleFunc("/", getPing)
	m.HandleFunc("/ws", hub.serveWS)
}

func getPing(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Pong")
}
