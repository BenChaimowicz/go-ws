package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	hub        *Hub
	egress     chan Event
}

func NewClient(conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		connection: conn,
		hub:        hub,
		egress:     make(chan Event),
	}
}

func (c *Client) readMessages() {
	defer func() {
		c.hub.removeClient(c)
	}()

	for {
		_, payload, err := c.connection.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println(err)
			}
			break
		}

		var request Event

		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("Error parsing event: %v", err)
			break
		}

		if err := c.hub.routeEvent(request, c); err != nil {
			log.Println("Error while routing message: ", err)
		}

		// log.Printf("Type: %v; Message: %v", msgType, string(payload))
	}
}

func (c *Client) writeMessages() {
	defer func() {
		c.hub.removeClient(c)
	}()
	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("Connection close: ", err)
				}
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
				return
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Failed to send message: %v", err)
			}
			log.Println("Message sent")
		}

	}
}
