package main

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var connectionUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Hub struct {
	clients  ClientList
	handlers map[string]EventHandler
	sync.RWMutex
}

func NewHub() *Hub {
	h := &Hub{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
	}
	h.setupHandlers()
	return h
}

func (h *Hub) serveWS(w http.ResponseWriter, r *http.Request) {
	conn, err := connectionUpgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}
	client := NewClient(conn, h)
	h.addClient(client)

	go client.readMessages()
	go client.writeMessages()
}

func (h *Hub) addClient(c *Client) {
	h.Lock()
	defer h.Unlock()

	h.clients[c] = true
}

func (h *Hub) removeClient(c *Client) {
	h.Lock()
	defer h.Unlock()

	if _, exists := h.clients[c]; exists {
		c.connection.Close()
		delete(h.clients, c)
	}
}

func (h *Hub) setupHandlers() {
	h.handlers[EventSendMessage] = SendMessage
}

func (h *Hub) routeEvent(event Event, c *Client) error {
	if handler, ok := h.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("invalid event type")
	}
}

func SendMessage(event Event, c *Client) error {
	log.Println(event)
	return nil
}
